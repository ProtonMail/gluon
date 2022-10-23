package logging

import (
	"context"
	"fmt"
	"runtime"
	"runtime/pprof"
	"strconv"
)

func GoAnnotate(ctx context.Context, fn func(context.Context), labelMap ...map[string]any) {
	go pprof.Do(ctx, getLabels(labelMap...), fn)
}

func DoAnnotate(ctx context.Context, fn func(context.Context), labelMap ...map[string]any) {
	pprof.Do(ctx, getLabels(labelMap...), fn)
}

func getLabels(labelMap ...map[string]any) pprof.LabelSet {
	// Get the caller's stack frame.
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		panic("failed to get caller's stack frame")
	}

	// Get the function name.
	fnName := runtime.FuncForPC(pc).Name()

	// Create the labels to annotate the goroutines with.
	labels := []string{"fn", fnName, "file", file, "line", strconv.Itoa(line)}

	// Add additional labels.
	for _, labelMap := range labelMap {
		for key, val := range labelMap {
			labels = append(labels, key, fmt.Sprintf("%v", val))
		}
	}

	return pprof.Labels(labels...)
}
