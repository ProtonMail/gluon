package observability

import "context"

type observabilitySenderKeyType struct{}

var observabilitySenderKeyVal observabilitySenderKeyType

func NewContextWithObservabilitySender(ctx context.Context, sender Sender) context.Context {
	return context.WithValue(ctx, observabilitySenderKeyVal, sender)
}

func getObservabilitySenderFromContext(ctx context.Context) (Sender, bool) {
	v := ctx.Value(observabilitySenderKeyVal)
	if v == nil {
		return nil, false
	}

	sender, ok := v.(Sender)
	return sender, ok
}
