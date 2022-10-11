package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/emersion/go-imap/client"
)

var serverUrl = flag.String("server", "127.0.0.1:1143", "IMAP server address:port")
var userName = flag.String("user-name", "user", "IMAP user name")
var userPassword = flag.String("user-pwd", "password", "IMAP user password")
var mbox = flag.String("mbox", "INBOX", "IMAP mailbox to append to")

func main() {
	flag.Parse()
	flag.Usage = func() {
		fmt.Printf("Usage %v [options] file0 ... fileN\n", os.Args[0])
		fmt.Printf("\nOptions:\n")
		flag.PrintDefaults()
	}

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		return
	}

	client, err := client.Dial(*serverUrl)
	if err != nil {
		panic(fmt.Errorf("failed to connect to server: %w", err))
	}

	defer func() {
		if err := client.Logout(); err != nil {
			panic(err)
		}
	}()

	if err := client.Login(*userName, *userPassword); err != nil {
		panic(fmt.Errorf("failed to login to server: %w", err))
	}

	for _, v := range args {
		fileData, err := os.ReadFile(v)
		if err != nil {
			panic(fmt.Errorf("failed to read file:%v - %w", v, err))
		}

		if err := client.Append(*mbox, []string{}, time.Now(), bytes.NewReader(fileData)); err != nil {
			panic(fmt.Errorf("failed to upload file:%v - %w", v, err))
		}
	}
}
