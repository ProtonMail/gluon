package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"path/filepath"
	"time"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"

	"entgo.io/ent/dialect"
	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/ProtonMail/gluon/store"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

// LocalServer runs a gluon server in the same process as the benchmark process.
type LocalServer struct {
	server   *gluon.Server
	listener net.Listener
}

func (l *LocalServer) Address() net.Addr {
	return l.listener.Addr()
}

func (l *LocalServer) Close(ctx context.Context) error {
	return l.server.Close(ctx)
}

type LocalServerBuilder struct{}

func (*LocalServerBuilder) New(ctx context.Context, serverPath string, profiler profiling.CmdProfilerBuilder) (Server, error) {
	loggerIn := logrus.StandardLogger().WriterLevel(logrus.TraceLevel)
	loggerOut := logrus.StandardLogger().WriterLevel(logrus.TraceLevel)

	var opts []gluon.Option

	opts = append(opts, gluon.WithLogger(loggerIn, loggerOut))
	opts = append(opts, gluon.WithCmdProfiler(profiler))

	server, err := gluon.New(serverPath, opts...)

	if err != nil {
		return nil, err
	}

	if err := addUser(ctx, server, serverPath, []string{utils.UserName}, utils.UserPassword); err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", "localhost:1143")
	if err != nil {
		return nil, err
	}

	go func() {
		for err := range server.Serve(ctx, listener) {
			logrus.WithError(err).Error("Error while serving")
		}
	}()

	return &LocalServer{
		server:   server,
		listener: listener,
	}, nil
}

func addUser(ctx context.Context, server *gluon.Server, path string, addresses []string, password string) error {
	hash := sha256.Sum256([]byte(addresses[0]))
	id := hex.EncodeToString(hash[:])
	connector := connector.NewDummy(
		addresses,
		password,
		time.Second,
		imap.NewFlagSet(`\Answered`, `\Seen`, `\Flagged`, `\Deleted`),
		imap.NewFlagSet(`\Answered`, `\Seen`, `\Flagged`, `\Deleted`),
		imap.NewFlagSet(),
	)

	storePath := filepath.Join(path, id+".store")
	dbPath := filepath.Join(path, id+".db")

	if *flags.VerboseFlag {
		fmt.Printf("Adding user ID=%v\n  StorePath:'%v'\n  DBPath:'%v'\n", id, storePath, dbPath)
	}

	store, err := store.NewOnDiskStore(storePath, []byte(utils.UserPassword))
	if err != nil {
		return err
	}

	_, err = server.AddUser(
		ctx,
		connector,
		store,
		dialect.SQLite,
		fmt.Sprintf("file:%v?cache=shared&_fk=1", dbPath))
	if err != nil {
		return err
	}

	if err := connector.Sync(context.Background()); err != nil {
		return err
	}

	return nil
}
