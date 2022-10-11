package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/pkg/profile"
	"github.com/sirupsen/logrus"
)

var (
	cpuProfileFlag   = flag.Bool("profile-cpu", false, "Enable CPU profiling.")
	memProfileFlag   = flag.Bool("profile-mem", false, "Enable Memory profiling.")
	blockProfileFlag = flag.Bool("profile-lock", false, "Enable lock profiling.")
	profilePathFlag  = flag.String("profile-path", "", "Path where to write profile data.")
)

func main() {
	ctx := context.Background()

	flag.Parse()

	if *cpuProfileFlag {
		p := profile.Start(profile.CPUProfile, profile.ProfilePath(*profilePathFlag))
		defer p.Stop()
	}

	if *memProfileFlag {
		p := profile.Start(profile.MemProfile, profile.MemProfileAllocs, profile.ProfilePath(*profilePathFlag))
		defer p.Stop()
	}

	if *blockProfileFlag {
		p := profile.Start(profile.BlockProfile, profile.ProfilePath(*profilePathFlag))
		defer p.Stop()
	}

	if level, err := logrus.ParseLevel(os.Getenv("GLUON_LOG_LEVEL")); err == nil {
		logrus.SetLevel(level)
	}

	server, err := gluon.New(
		gluon.WithLogger(
			logrus.StandardLogger().WriterLevel(logrus.TraceLevel),
			logrus.StandardLogger().WriterLevel(logrus.TraceLevel),
		),
		gluon.WithDataDir(os.Getenv("GLUON_DIR")),
	)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create server")
	}

	defer server.Close(ctx)

	nUsers := 2
	if envUsers, err := strconv.Atoi(os.Getenv("GLUON_USER_COUNT")); err == nil {
		nUsers = envUsers
	}

	for i := 1; i <= nUsers; i++ {
		if err := addUser(ctx, server, []string{
			fmt.Sprintf("user%d@example.com", i),
			fmt.Sprintf("alias%d@example.com", i),
		}, []byte("pass")); err != nil {
			logrus.WithError(err).Fatal("Failed to add user")
		}
	}

	host := "localhost:1143"
	if envHost := os.Getenv("GLUON_HOST"); envHost != "" {
		host = envHost
	}

	listener, err := net.Listen("tcp", host)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to listen")
	}

	logrus.Infof("Server is listening on %v", listener.Addr())

	if err := server.Serve(ctx, listener); err != nil {
		logrus.WithError(err).Fatal("Failed to serve")
	}

	for err := range server.GetErrorCh() {
		logrus.WithError(err).Error("Error while serving")
	}

	if err := listener.Close(); err != nil {
		logrus.WithError(err).Error("Failed to close listener")
	}
}

func addUser(ctx context.Context, server *gluon.Server, addresses []string, password []byte) error {
	connector := connector.NewDummy(
		addresses,
		password,
		time.Second,
		imap.NewFlagSet(`\Answered`, `\Seen`, `\Flagged`, `\Deleted`),
		imap.NewFlagSet(`\Answered`, `\Seen`, `\Flagged`, `\Deleted`),
		imap.NewFlagSet(),
	)

	userID, err := server.AddUser(
		ctx,
		connector,
		password,
	)
	if err != nil {
		return err
	}

	if err := connector.Sync(ctx); err != nil {
		return err
	}

	logrus.WithField("userID", userID).Info("User added to server")

	return nil
}
