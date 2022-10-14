package response

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
	"golang.org/x/exp/slices"
)

type itemCapability struct {
	caps []imap.Capability
}

func ItemCapability(caps ...imap.Capability) *itemCapability {
	return &itemCapability{
		caps: caps,
	}
}

func (r *itemCapability) Strings() (raw string, _ string) {
	var caps []string

	for _, capability := range r.caps {
		caps = append(caps, string(capability))
	}

	slices.Sort(caps)

	raw = fmt.Sprintf("CAPABILITY %v", join(caps))

	return raw, raw
}
