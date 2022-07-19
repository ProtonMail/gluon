package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"entgo.io/ent/dialect"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"

	"net"
	"path/filepath"
	"time"

	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/store"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

// newServer creates a new server and returns the server instance and address.
func newServer(ctx context.Context, path string, options ...gluon.Option) (*gluon.Server, string, error) {
	loggerIn := logrus.StandardLogger().WriterLevel(logrus.TraceLevel)
	loggerOut := logrus.StandardLogger().WriterLevel(logrus.TraceLevel)

	var opts []gluon.Option

	opts = append(opts, gluon.WithLogger(loggerIn, loggerOut))
	opts = append(opts, options...)

	server, err := gluon.New(path, opts...)

	if err != nil {
		return nil, "", err
	}

	if err := addUser(ctx, server, path, []string{utils.UserName}, utils.UserPassword); err != nil {
		return nil, "", err
	}

	listener, err := net.Listen("tcp", "localhost:1143")
	if err != nil {
		return nil, "", err
	}

	go func() {
		for err := range server.Serve(ctx, listener) {
			logrus.WithError(err).Error("Error while serving")
		}
	}()

	return server, listener.Addr().String(), nil
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

	if *verboseFlag {
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

	if *verboseFlag {
		fmt.Printf("Starting Connector Sync\n")
	}

	if err := connector.Sync(context.Background()); err != nil {
		if *verboseFlag {
			fmt.Printf("Error during Connector Sync\n")
		}

		return err
	}

	if *verboseFlag {
		fmt.Printf("Finished Connector Sync\n")
	}

	return nil
}
