package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"entgo.io/ent/dialect"
	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/store"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

func main() {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to make temporary directory")
	}

	server := gluon.New(filepath.Join(dir, "server"))

	connector := connector.NewDummy(
		[]string{"user@example.com", "alias@example.com"},
		"password",
		time.Second,
		imap.NewFlagSet(),
		imap.NewFlagSet(),
		imap.NewFlagSet(),
	)

	store, err := store.NewOnDiskStore(filepath.Join(dir, "store"), []byte("passphrase"))
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create store")
	}

	userID, err := server.AddUser(
		connector,
		store,
		dialect.SQLite,
		fmt.Sprintf("file:%v?cache=shared&_fk=1", filepath.Join(dir, fmt.Sprintf("%v.db", "userID"))),
	)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to add user")
	}

	logrus.WithField("userID", userID).Info("User added to server")

	listener, err := net.Listen("tcp", ":1143")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to listen")
	}

	for err := range server.Serve(context.Background(), listener) {
		logrus.WithError(err).Error("Error while serving")
	}
}
