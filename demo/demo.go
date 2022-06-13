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
	server, err := gluon.New(temp())
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create server")
	}

	if err := addUser(server, []string{"user1@example.com", "alias1@example.com"}, "password1"); err != nil {
		logrus.WithError(err).Fatal("Failed to add user")
	}

	if err := addUser(server, []string{"user2@example.com", "alias2@example.com"}, "password2"); err != nil {
		logrus.WithError(err).Fatal("Failed to add user")
	}

	listener, err := net.Listen("tcp", "localhost:1143")
	if err != nil {
		logrus.WithError(err).Fatal("Failed to listen")
	}

	logrus.Infof("Server is listening on %v", listener.Addr())

	for err := range server.Serve(context.Background(), listener) {
		logrus.WithError(err).Error("Error while serving")
	}
}

func addUser(server *gluon.Server, addresses []string, password string) error {
	connector := connector.NewDummy(
		addresses,
		password,
		time.Second,
		imap.NewFlagSet(),
		imap.NewFlagSet(),
		imap.NewFlagSet(),
	)

	store, err := store.NewOnDiskStore(temp(), []byte("passphrase"))
	if err != nil {
		return err
	}

	userID, err := server.AddUser(
		connector,
		store,
		dialect.SQLite,
		fmt.Sprintf("file:%v?cache=shared&_fk=1", filepath.Join(temp(), fmt.Sprintf("%v.db", "userID"))),
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
