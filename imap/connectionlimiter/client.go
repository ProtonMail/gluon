package connectionlimiter

import (
	"strings"

	"github.com/ProtonMail/gluon/imap"
)

type Client string

const (
	ClientAppleMail   Client = "apple-mail"
	ClientOutlook     Client = "outlook"
	ClientThunderbird Client = "thunderbird"
	ClientUnknown     Client = "unknown"
)

func normalizeClientKey(id imap.IMAPID) Client {
	name := strings.TrimSpace(strings.ToLower(id.Name))
	switch {
	case strings.Contains(name, "outlook"):
		return ClientOutlook
	case strings.Contains(name, "thunderbird"):
		return ClientThunderbird
	case strings.Contains(name, "mac") && strings.Contains(name, "mail"):
		return ClientAppleMail
	case strings.Contains(name, "mac") && strings.Contains(name, "notes"):
		return ClientUnknown
	default:
		return ClientUnknown
	}
}
