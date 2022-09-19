package gluon_benchmarks

import (
	"context"
	"flag"
	"math/rand"
	"time"

	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/timing"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/hash"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var (
	syncMessageCountFlag = flag.Uint("sync-msg-count", 1000, "Number of messages to sync.")
	syncMBoxCountFlag    = flag.Uint("sync-mbox-count", 1, "Number of mailboxes to sync.")
)

type Sync struct {
	connector utils.ConnectorImpl
	server    *gluon.Server
	mailboxes []imap.LabelID
}

func NewSync() benchmark.Benchmark {
	return &Sync{}
}

func (s *Sync) Name() string {
	return "gluon-sync"
}

func (s *Sync) Setup(ctx context.Context, benchmarkDir string) error {
	loggerIn := logrus.StandardLogger().WriterLevel(logrus.TraceLevel)
	loggerOut := logrus.StandardLogger().WriterLevel(logrus.TraceLevel)

	opts := []gluon.Option{
		gluon.WithDataDir(benchmarkDir),
		gluon.WithLogger(loggerIn, loggerOut),
	}

	server, err := gluon.New(opts...)
	if err != nil {
		return err
	}

	s.server = server

	connector, err := s.setupConnector(ctx)
	if err != nil {
		return err
	}

	s.connector = connector

	return nil
}

func (s *Sync) setupConnector(ctx context.Context) (utils.ConnectorImpl, error) {
	c, err := utils.NewConnector(*flags.Connector)
	if err != nil {
		return nil, err
	}

	mboxIDs := make([]imap.LabelID, 0, *syncMBoxCountFlag)

	for i := uint(0); i < *syncMBoxCountFlag; i++ {
		mbox, err := c.Connector().CreateLabel(ctx, []string{uuid.NewString()})
		if err != nil {
			return nil, err
		}

		mboxIDs = append(mboxIDs, mbox.ID)
	}

	messages := [][]byte{
		[]byte(utils.MessageEmbedded),
		[]byte(utils.MessageMultiPartMixed),
		[]byte(utils.MessageAfterNoonMeeting),
	}

	parsedMessages := make([]*imap.ParsedMessage, len(messages))

	for i, m := range messages {
		pmsg, err := imap.NewParsedMessage(m)
		if err != nil {
			return nil, err
		}

		parsedMessages[i] = pmsg
	}

	flagSet := imap.NewFlagSet("\\Recent", "\\Draft", "\\Foo")

	s.mailboxes = make([]imap.LabelID, 0, len(mboxIDs))

	for _, mboxID := range mboxIDs {
		for i := uint(0); i < *syncMessageCountFlag; i++ {
			randIndex := rand.Intn(len(messages))
			if _, err := c.Connector().CreateMessage(ctx, mboxID, messages[randIndex], parsedMessages[randIndex], flagSet, time.Now()); err != nil {
				return nil, err
			}
		}

		s.mailboxes = append(s.mailboxes, mboxID)
	}

	if _, err = s.server.AddUser(
		ctx,
		c.Connector(),
		hash.SHA256([]byte(*flags.UserPassword)),
	); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *Sync) Run(ctx context.Context) (*reporter.BenchmarkRun, error) {
	timer := timing.Timer{}

	timer.Start()

	err := s.connector.Sync(ctx)

	timer.Stop()

	if err != nil {
		return nil, err
	}

	return reporter.NewBenchmarkRunSingle(timer.Elapsed(), nil), nil
}

func (s *Sync) TearDown(ctx context.Context) error {
	for _, id := range s.mailboxes {
		if err := s.connector.Connector().DeleteLabel(ctx, id); err != nil {
			return err
		}
	}

	if s.server != nil {
		if err := s.server.Close(ctx); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	benchmark.RegisterBenchmark(NewSync())
}
