package imap

type Capability string

const (
	IMAP4rev1 Capability = `IMAP4rev1`
	StartTLS  Capability = `STARTTLS`
	IDLE      Capability = `IDLE`
	UNSELECT  Capability = `UNSELECT`
	UIDPLUS   Capability = `UIDPLUS`
	MOVE      Capability = `MOVE`
)

func IsCapabilityAvailableBeforeAuth(c Capability) bool {
	switch c {
	case IMAP4rev1, StartTLS, IDLE:
		return true
	case UNSELECT, UIDPLUS, MOVE:
		return false
	}

	return false
}
