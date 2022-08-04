package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"entgo.io/ent/dialect"
	"fmt"
	"net"
	"path/filepath"

	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
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

	if err := addUser(ctx, server, serverPath); err != nil {
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

func addUser(ctx context.Context, server *gluon.Server, path string) error {
	hash := sha256.Sum256([]byte(*flags.UserName))
	id := hex.EncodeToString(hash[:])

	c, err := utils.NewConnector(*flags.Connector)
	if err != nil {
		return nil
	}

	storePath := filepath.Join(path, id+".store")
	dbPath := filepath.Join(path, id+".db")

	if *flags.Verbose {
		fmt.Printf("Adding user ID=%v\n  BenchPath:'%v'\n  DBPath:'%v'\n", id, storePath, dbPath)
	}

	store, err := store.NewOnDiskStore(storePath, []byte(*flags.UserPassword))
	if err != nil {
		return err
	}

	_, err = server.AddUser(
		ctx,
		c.Connector(),
		store,
		dialect.SQLite,
		fmt.Sprintf("file:%v?cache=shared&_fk=1", dbPath))
	if err != nil {
		return err
	}

	if err := c.Sync(context.Background()); err != nil {
		return err
	}

	return nil
}
