package response

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
)

type flags struct {
	flags imap.FlagSet
}

func Flags() *flags {
	return &flags{flags: imap.NewFlagSet()}
}

func (r *flags) WithFlags(fs imap.FlagSet) *flags {
	r.flags = r.flags.AddFlagSet(fs)
	return r
}

func (r *flags) Send(s Session) error {
	return s.WriteResponse(r.String())
}

func (r *flags) String() string {
	return fmt.Sprintf("* FLAGS (%v)", join(r.flags.ToSlice()))
}
