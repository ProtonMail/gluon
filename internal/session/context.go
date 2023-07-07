package session

import (
	"context"
	"fmt"
	"time"
)

type startTimeType struct{}

var startTimeKey startTimeType

func withStartTime(parent context.Context, startTime time.Time) context.Context {
	return context.WithValue(parent, startTimeKey, startTime)
}

func startTimeFromContext(ctx context.Context) (time.Time, bool) {
	startTime, ok := ctx.Value(startTimeKey).(time.Time)
	return startTime, ok
}

func okMessage(ctx context.Context) string {
	if startTime, ok := startTimeFromContext(ctx); ok {
		elapsed := time.Since(startTime)
		microSec := elapsed.Microseconds()

		return fmt.Sprintf("command completed in %v microsec.", microSec)
	}

	return ""
}
