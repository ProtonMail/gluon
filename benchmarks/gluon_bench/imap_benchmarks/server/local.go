package server

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net"

	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
	"github.com/ProtonMail/gluon/profiling"
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
	opts = append(opts, gluon.WithDataDir(serverPath))

	server, err := gluon.New(opts...)
	if err != nil {
		return nil, err
	}

	if err := addUser(ctx, server); err != nil {
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

func addUser(ctx context.Context, server *gluon.Server) error {
	c, err := utils.NewConnector(*flags.Connector)
	if err != nil {
		return err
	}

	encryptionBytes := sha256.Sum256([]byte(*flags.UserPassword))

	if userID, err := server.AddUser(
		ctx,
		c.Connector(),
		encryptionBytes[:]); err != nil {
		return err
	} else if *flags.Verbose {
		fmt.Printf("Adding user ID=%v\n", userID)
	}

	if err := c.Sync(context.Background()); err != nil {
		return err
	}

	return nil
}
