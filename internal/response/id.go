package response

import (
	"github.com/ProtonMail/gluon/imap"
)

type idResponse struct {
	imap.IMAPID
}

func ID(id imap.IMAPID) *idResponse {
	return &idResponse{
		IMAPID: id,
	}
}

func (id *idResponse) Strings() (raw string, _ string) {
	raw = "* ID " + id.IMAPID.String()
	return raw, raw
}

func (r *idResponse) Send(session Session) error {
	return session.WriteResponse(r)
}
