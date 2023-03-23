package backend

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/db/ent/mailbox"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/limits"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/store"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// maxLoginAttempts is the maximum number of permitted login attempts before the user is jailed.
const maxLoginAttempts = 3

type Backend struct {
	// dataDir is the directory in which backend files should be stored.
	dataDir string

	// databaseDir is the directory in which database files should be stored.
	databaseDir string

	// delim is the server's path delim.
	delim string

	// users holds all registered backend users.
	users     map[string]*user
	usersLock sync.Mutex

	// storeBuilder builds stores for the backend users.
	storeBuilder store.Builder

	// loginJailTime is the time a user is jailed after too many failed login attempts.
	loginJailTime time.Duration

	// loginErrorCount holds the number of failed login attempts for each user.
	loginErrorCount int32
	loginLock       sync.Mutex
	loginWG         sync.WaitGroup

	imapLimits limits.IMAP

	panicHandler queue.PanicHandler
}

func New(dataDir, databaseDir string, storeBuilder store.Builder, delim string, loginJailTime time.Duration, imapLimits limits.IMAP, panicHandler queue.PanicHandler) (*Backend, error) {
	return &Backend{
		dataDir:       dataDir,
		databaseDir:   databaseDir,
		delim:         delim,
		users:         make(map[string]*user),
		storeBuilder:  storeBuilder,
		loginJailTime: loginJailTime,
		imapLimits:    imapLimits,
	}, nil
}

func (b *Backend) NewUserID() string {
	return uuid.NewString()
}

func (b *Backend) GetDelimiter() string {
	return b.delim
}

// AddUser adds a new user to the backend.
// It returns true if the user's database was created, false if it already existed.
func (b *Backend) AddUser(ctx context.Context, userID string, conn connector.Connector, passphrase []byte, uidValidityGenerator imap.UIDValidityGenerator) (bool, error) {
	b.usersLock.Lock()
	defer b.usersLock.Unlock()

	storeBuilder, err := b.storeBuilder.New(b.getStoreDir(), userID, passphrase)
	if err != nil {
		return false, err
	}

	db, isNew, err := db.NewDB(b.getDBDir(), userID)
	if err != nil {
		if err := storeBuilder.Close(); err != nil {
			logrus.WithError(err).Error("Failed to close store builder")
		}

		return false, err
	}

	user, err := newUser(ctx, userID, db, conn, storeBuilder, b.delim, b.imapLimits, uidValidityGenerator, b.panicHandler)
	if err != nil {
		return false, err
	}

	b.users[userID] = user

	return isNew, nil
}

func (b *Backend) RemoveUser(ctx context.Context, userID string, removeFiles bool) error {
	b.usersLock.Lock()
	defer b.usersLock.Unlock()

	user, ok := b.users[userID]
	if !ok {
		return ErrNoSuchUser
	}

	if err := user.close(ctx); err != nil {
		reporter.MessageWithContext(ctx,
			"Failed to close user from Backend.RemoveUser()",
			reporter.Context{"error": err},
		)

		return fmt.Errorf("failed to close backend user: %w", err)
	}

	delete(b.users, userID)

	if removeFiles {
		if err := b.storeBuilder.Delete(b.getStoreDir(), userID); err != nil {
			return err
		}

		if err := db.DeleteDB(b.getDBDir(), userID); err != nil {
			return err
		}
	}

	return nil
}

func (b *Backend) GetMailboxMessageCounts(ctx context.Context, userID string) (map[imap.MailboxID]int, error) {
	b.usersLock.Lock()
	defer b.usersLock.Unlock()

	user, ok := b.users[userID]
	if !ok {
		return nil, ErrNoSuchUser
	}

	return db.ReadResult(ctx, user.db, func(ctx context.Context, c *ent.Client) (map[imap.MailboxID]int, error) {
		counts := make(map[imap.MailboxID]int)

		mailboxes, err := c.Mailbox.Query().Select(mailbox.FieldRemoteID).All(ctx)
		if err != nil {
			return nil, err
		}

		for _, mailbox := range mailboxes {
			messageCount, err := mailbox.QueryUIDs().Count(ctx)
			if err != nil {
				return nil, err
			}

			counts[mailbox.RemoteID] = messageCount
		}

		return counts, nil
	})
}

func (b *Backend) GetState(ctx context.Context, username string, password []byte, sessionID int) (*state.State, error) {
	b.usersLock.Lock()
	defer b.usersLock.Unlock()

	userID, err := b.getUserID(ctx, username, password)
	if err != nil {
		// todo filter on error and track ErrLoginBlocked to notify the connector
		return nil, err
	}

	state, err := b.users[userID].newState()
	if err != nil {
		return nil, err
	}

	logrus.
		WithField("userID", userID).
		WithField("username", username).
		WithField("stateID", state.StateID).
		Debug("Created new IMAP state")

	return state, nil
}

func (b *Backend) ReleaseState(ctx context.Context, st *state.State) error {
	b.usersLock.Lock()
	defer b.usersLock.Unlock()

	userID := st.UserID()
	user, ok := b.users[userID]

	if !ok {
		return ErrNoSuchUser
	}

	return user.removeState(ctx, st)
}

func (b *Backend) Close(ctx context.Context) error {
	b.usersLock.Lock()
	defer b.usersLock.Unlock()

	for userID, user := range b.users {
		if err := user.close(ctx); err != nil {
			reporter.MessageWithContext(ctx,
				"Failed to close user from Backend.Close()",
				reporter.Context{"error": err},
			)

			return fmt.Errorf("failed to close backend user (%v): %w", userID, err)
		}

		delete(b.users, userID)
	}

	logrus.Debug("Backend was closed")

	return nil
}

func (b *Backend) getUserID(ctx context.Context, username string, password []byte) (string, error) {
	b.loginLock.Lock()
	defer b.loginLock.Unlock()

	b.loginWG.Wait()

	for _, user := range b.users {
		if user.connector.Authorize(username, password) {
			atomic.StoreInt32(&b.loginErrorCount, 0)
			return user.userID, nil
		}
	}

	if count := atomic.AddInt32(&b.loginErrorCount, 1); count == maxLoginAttempts {
		b.loginWG.Add(1)

		time.AfterFunc(b.loginJailTime, func() {
			defer b.loginWG.Done()
			atomic.StoreInt32(&b.loginErrorCount, 0)
		})

		return "", ErrLoginBlocked
	}

	return "", ErrNoSuchUser
}

func (b *Backend) getStoreDir() string {
	return b.dataDir
}

func (b *Backend) getDBDir() string {
	return b.databaseDir
}
