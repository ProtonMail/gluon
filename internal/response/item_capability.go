package response

import (
	"fmt"
	"slices"

	"github.com/ProtonMail/gluon/imap"
)

type itemCapability struct {
	caps []imap.Capability
}

func ItemCapability(caps ...imap.Capability) *itemCapability {
	return &itemCapability{
		caps: caps,
	}
}

func (r *itemCapability) String() string {
	var caps []string

	for _, capability := range r.caps {
		caps = append(caps, string(capability))
	}

	slices.Sort(caps)

	return fmt.Sprintf("CAPABILITY %v", join(caps))
}
