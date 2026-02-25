package response

import (
	"fmt"
	"strings"
)

type quotaroot struct {
	mailbox   string
	rootNames []string
}

func QuotaRoot(mailbox string, rootNames []string) *quotaroot {
	return &quotaroot{
		mailbox:   mailbox,
		rootNames: rootNames,
	}
}

func (r *quotaroot) Send(s Session) error {
	return s.WriteResponse(r.String())
}

func (r *quotaroot) String() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("* QUOTAROOT %v", r.mailbox))

	for _, name := range r.rootNames {
		if name == "" {
			parts = append(parts, `""`)
		} else {
			parts = append(parts, fmt.Sprintf(`"%v"`, name))
		}
	}

	return strings.Join(parts, " ")
}
