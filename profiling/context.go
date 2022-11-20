package profiling

import "context"

type withProfilerType struct{}

var withProfilerKey withProfilerType

func Start(ctx context.Context, cmdType int) {
	if profiler, ok := ctx.Value(withProfilerKey).(CmdProfiler); ok {
		profiler.Start(cmdType)
	}
}

func Stop(ctx context.Context, cmdType int) {
	if profiler, ok := ctx.Value(withProfilerKey).(CmdProfiler); ok {
		profiler.Stop(cmdType)
	}
}

func WithProfiler(ctx context.Context, profiler CmdProfiler) context.Context {
	return context.WithValue(ctx, withProfilerKey, profiler)
}
