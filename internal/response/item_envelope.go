package response

import "fmt"

type itemEnvelope struct {
	envelope string
}

func ItemEnvelope(envelope string) *itemEnvelope {
	return &itemEnvelope{
		envelope: envelope,
	}
}

func (r *itemEnvelope) String() string {
	return fmt.Sprintf("ENVELOPE %v", r.envelope)
}
