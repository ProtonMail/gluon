package command

import (
	"fmt"

	"github.com/ProtonMail/gluon/rfcparser"
)

type Capability struct{}

func (l Capability) String() string {
	return fmt.Sprintf("CAPABILITY")
}

func (l Capability) SanitizedString() string {
	return l.String()
}

type CapabilityCommandParser struct{}

func (CapabilityCommandParser) FromParser(p *rfcparser.Parser) (Payload, error) {
	return &Capability{}, nil
}
