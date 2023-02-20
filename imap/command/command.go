package command

import (
	"encoding/base64"
	"fmt"

	"github.com/ProtonMail/gluon/internal/hash"
)

type Payload interface {
	String() string

	// SanitizedString should return the command payload with all the sensitive information stripped out.
	SanitizedString() string
}

func sanitizeString(s string) string {
	return base64.StdEncoding.EncodeToString(hash.SHA256([]byte(s)))
}

type Command struct {
	Tag     string
	Payload Payload
}

func (c Command) String() string {
	return fmt.Sprintf("%v %v", c.Tag, c.Payload.String())
}

func (c Command) SanitizedString() string {
	return fmt.Sprintf("%v %v", c.Tag, c.Payload.SanitizedString())
}
