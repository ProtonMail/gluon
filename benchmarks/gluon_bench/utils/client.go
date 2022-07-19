package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-imap/client"
)

func NewClient(addr string) (*client.Client, error) {
	client, err := client.Dial(addr)
	if err != nil {
		return nil, err
	}

	if err := client.Login(UserName, UserPassword); err != nil {
		return nil, err
	}

	return client, nil
}

func AppendToMailbox(cl *client.Client, mailboxName string, literal string, time time.Time, flags ...string) error {
	return cl.Append(mailboxName, flags, time, strings.NewReader(literal))
}

func CloseClient(cl *client.Client) {
	if err := cl.Logout(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to close client: %v\n", err)
	}
}
