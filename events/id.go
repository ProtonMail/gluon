package events

import "github.com/ProtonMail/gluon/imap"

type IMAPID struct {
	eventBase

	SessionID int
	IMAPID    imap.IMAPID
}
