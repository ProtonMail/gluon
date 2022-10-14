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
	r.flags.AddFlagSetToSelf(fs)
	return r
}

func (r *flags) Send(s Session) error {
	return s.WriteResponse(r)
}

func (r *flags) Strings() (raw string, _ string) {
	raw = fmt.Sprintf("* FLAGS (%v)", join(r.flags.ToSlice()))
	return raw, raw
}
