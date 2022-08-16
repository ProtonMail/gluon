package main

import (
	"context"
	"flag"
	"net"
	"os"
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

	server, err := gluon.New(gluon.WithLogger(
		logrus.StandardLogger().WriterLevel(logrus.TraceLevel),
		logrus.StandardLogger().WriterLevel(logrus.TraceLevel),
	))
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create server")
	}

	defer server.Close(ctx)

	if err := addUser(ctx, server, []string{"user1@example.com", "alias1@example.com"}, "password1"); err != nil {
		logrus.WithError(err).Fatal("Failed to add user")
	}

	if err := addUser(ctx, server, []string{"user2@example.com", "alias2@example.com"}, "password2"); err != nil {
		logrus.WithError(err).Fatal("Failed to add user")
	}

	listener, err := net.Listen("tcp", "localhost:1143")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to listen")
	}

	logrus.Infof("Server is listening on %v", listener.Addr())

	for err := range server.Serve(ctx, listener) {
		logrus.WithError(err).Error("Error while serving")
	}
}

func addUser(ctx context.Context, server *gluon.Server, addresses []string, password string) error {
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
		[]byte(password),
	)
	if err != nil {
		return err
	}

	if err := connector.Sync(context.Background()); err != nil {
		return err
	}

	logrus.WithField("userID", userID).Info("User added to server")

	return nil
}

func temp() string {
	temp, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}

	return temp
}
