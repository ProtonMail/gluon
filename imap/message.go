package imap

import "time"

type Message struct {
	ID    string
	Flags FlagSet
	Date  time.Time
}

type Header []Field

type Field struct {
	Key, Value string
}

func (m *Message) HasFlag(wantFlag string) bool {
	return m.Flags.Contains(wantFlag)
}
