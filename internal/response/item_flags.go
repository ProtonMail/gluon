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

func (c *itemFlags) String() string {
	return fmt.Sprintf("FLAGS (%v)", join(c.flags.ToSlice()))
}

func (c *itemFlags) mergeWith(other Item) Item {
	_, ok := other.(*itemFlags)
	if !ok {
		return nil
	}

	return ItemFlags(c.flags.Clone())
}
