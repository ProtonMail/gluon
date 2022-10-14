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

func (r *itemEnvelope) Strings() (raw string, _ string) {
	raw = fmt.Sprintf("ENVELOPE %v", r.envelope)
	return raw, raw
}
