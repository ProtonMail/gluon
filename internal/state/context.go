package state

import "context"

type stateContextType struct{}

var stateContextKey stateContextType

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
