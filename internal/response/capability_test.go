package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/assert"
)

func TestCapabilityUntagged(t *testing.T) {
	assert.Equal(t, "* CAPABILITY IMAP4rev1", Capability().WithCapabilities(imap.IMAP4rev1).String(false))
}

func TestCapabilityExtras(t *testing.T) {
	assert.Equal(t, "* CAPABILITY IDLE IMAP4rev1", Capability().WithCapabilities(imap.IMAP4rev1, imap.IDLE).String(false))
}
