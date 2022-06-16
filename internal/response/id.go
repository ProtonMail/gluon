package response

import (
	"github.com/ProtonMail/gluon/imap"
)

type idResponse struct {
	imap.ID
}

func ID(id imap.ID) *idResponse {
	return &idResponse{
		ID: id,
	}
}

func (id *idResponse) String() string {
	return "* ID " + id.ID.String()
}

func (r *idResponse) Send(session Session) error {
	return session.WriteResponse(r.String())
}
