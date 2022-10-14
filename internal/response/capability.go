package response

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
	"golang.org/x/exp/slices"
)

type capability struct {
	caps []imap.Capability
}

func Capability() *capability {
	return &capability{}
}

func (r *capability) WithCapabilities(caps ...imap.Capability) *capability {
	r.caps = append(r.caps, caps...)
	return r
}

func (r *capability) Send(s Session) error {
	return s.WriteResponse(r)
}

func (r *capability) String(_ bool) string {
	var caps []string

	for _, capability := range r.caps {
		caps = append(caps, string(capability))
	}

	slices.Sort(caps)

	return fmt.Sprintf("* CAPABILITY %v", join(caps))
}
