package imap

type Capability string

const (
	IMAP4rev1 Capability = `IMAP4rev1`
	StartTLS  Capability = `STARTTLS`
	IDLE      Capability = `IDLE`
	UNSELECT  Capability = `UNSELECT`
	UIDPLUS   Capability = `UIDPLUS`
	MOVE      Capability = `MOVE`
	ID        Capability = `ID`
)

func IsCapabilityAvailableBeforeAuth(c Capability) bool {
	switch c {
	case IMAP4rev1, StartTLS, IDLE, ID:
		return true
	case UNSELECT, UIDPLUS, MOVE:
		return false
	}

	return false
}
