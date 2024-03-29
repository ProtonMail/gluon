package contexts

import (
	"context"
)

type handleUIDType struct{}

var handleUIDKey handleUIDType

// AsUID marks this context as handling a UID command.
// This modifies some backend behaviour (such as returning UID within FETCH responses).
func AsUID(parent context.Context) context.Context {
	return context.WithValue(parent, handleUIDKey, struct{}{})
}

func IsUID(ctx context.Context) bool {
	return ctx.Value(handleUIDKey) != nil
}

type handleCloseType struct{}

var handleCloseKey handleCloseType

// AsClose marks this context as handling a CLOSE command.
// This modifies some backend behaviour (such as not returning EXPUNGE responses).
func AsClose(parent context.Context) context.Context {
	return context.WithValue(parent, handleCloseKey, struct{}{})
}

func IsClose(ctx context.Context) bool {
	return ctx.Value(handleCloseKey) != nil
}

type handleSilentType struct{}

var handleSilentKey handleSilentType

// AsSilent marks this context as handling a silent STORE command.
// This modifies some backend behaviour (such as not returning EXPUNGE responses).
func AsSilent(parent context.Context) context.Context {
	return context.WithValue(parent, handleSilentKey, struct{}{})
}

func IsSilent(ctx context.Context) bool {
	return ctx.Value(handleSilentKey) != nil
}

type disableParallelismCtxType struct{}

var disableParallelismCtxKey disableParallelismCtxType

func IsParallelismDisabledCtx(ctx context.Context) bool {
	return ctx.Value(disableParallelismCtxKey) != nil
}

func NewDisableParallelismCtx(ctx context.Context, disabled bool) context.Context {
	if !disabled {
		return ctx
	}

	return context.WithValue(ctx, disableParallelismCtxKey, struct{}{})
}
