package state

import "context"

type stateContextType struct{}

var stateContextKey stateContextType

// NewStateContext will annotate a context object with the state's assigned ID. This can later be used
// to determine whether the current active call came from a state and which one.
func NewStateContext(ctx context.Context, s *State) context.Context {
	if s == nil {
		return ctx
	}

	return context.WithValue(ctx, stateContextKey, s.StateID)
}

func GetStateIDFromContext(ctx context.Context) (int, bool) {
	v := ctx.Value(stateContextKey)
	if v == nil {
		return 0, false
	}

	stateID, ok := v.(int)

	return stateID, ok
}
