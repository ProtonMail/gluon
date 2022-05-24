package session

import (
	"errors"

	"github.com/ProtonMail/gluon/imap"
)

var ErrFlagRecentIsReserved = errors.New(`system flag \Recent is reserved`)

// validateStoreFlags ensures that the given flags are valid for a STORE command and return them as an imap.FlagSet.
func validateStoreFlags(flags []string) (imap.FlagSet, error) {
	flagSet := imap.NewFlagSetFromSlice(flags)

	// As per RFC 3501, section 2.3.2, changing the \Recent flag is forbidden.
	if flagSet.Contains(imap.FlagRecent) {
		return nil, ErrFlagRecentIsReserved
	}

	return flagSet, nil
}
