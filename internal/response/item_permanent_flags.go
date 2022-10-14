package response

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
)

type itemPermanentFlags struct {
	flags imap.FlagSet
}

func ItemPermanentFlags(flags imap.FlagSet) *itemPermanentFlags {
	return &itemPermanentFlags{flags: flags}
}

func (c *itemPermanentFlags) String(_ bool) string {
	return fmt.Sprintf("PERMANENTFLAGS (%v)", join(c.flags.ToSlice()))
}
