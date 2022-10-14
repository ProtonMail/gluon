// Package response implements types used when sending IMAP responses back to clients.
package response

type Response interface {
	Send(Session) error
	String(isPrivateByDefault bool) string
}

type Session interface {
	WriteResponse(item Item) error
}

type mergeableResponse interface {
	mergeWith(Response) Response
	canSkip(Response) bool
}
