package response

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
)

type itemFlags struct {
	flags imap.FlagSet
}

func ItemFlags(flags imap.FlagSet) *itemFlags {
	return &itemFlags{flags: flags}
}

func (c *itemFlags) Strings() (raw string, filtered string) {
	raw = fmt.Sprintf("FLAGS (%v)", join(c.flags.ToSlice()))
	return raw, raw
}

func (c *itemFlags) mergeWith(other Item) Item {
	_, ok := other.(*itemFlags)
	if !ok {
		return nil
	}

	return ItemFlags(c.flags.Clone())
}
