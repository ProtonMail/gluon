package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/assert"
)

func TestCapabilityUntagged(t *testing.T) {
	raw, filtered := Capability().WithCapabilities(imap.IMAP4rev1).Strings()
	assert.Equal(t, "* CAPABILITY IMAP4rev1", raw)
	assert.Equal(t, "* CAPABILITY IMAP4rev1", filtered)
}

func TestCapabilityExtras(t *testing.T) {
	raw, filtered := Capability().WithCapabilities(imap.IMAP4rev1, imap.IDLE).Strings()
	assert.Equal(t, "* CAPABILITY IDLE IMAP4rev1", raw)
	assert.Equal(t, "* CAPABILITY IDLE IMAP4rev1", filtered)
}
