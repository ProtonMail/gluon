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

func (id *idResponse) String(_ bool) string {
	return "* ID " + id.IMAPID.String()
}

func (r *idResponse) Send(session Session) error {
	return session.WriteResponse(r)
}
