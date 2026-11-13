package response

import (
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/imap"
)

type quota struct {
	root *imap.QuotaRoot
}

func Quota(root *imap.QuotaRoot) *quota {
	return &quota{root: root}
}

func (r *quota) Send(s Session) error {
	return s.WriteResponse(r.String())
}

func (r *quota) String() string {
	var resources []string

	for _, res := range r.root.Resources {
		resources = append(resources, fmt.Sprintf("%v %v %v", res.ResourceName, res.Usage, res.Limit))
	}

	return fmt.Sprintf(`* QUOTA %v (%v)`, quoteQuotaRoot(r.root.RootName), strings.Join(resources, " "))
}

func quoteQuotaRoot(name string) string {
	if name == "" {
		return `""`
	}

	return fmt.Sprintf(`"%v"`, name)
}
