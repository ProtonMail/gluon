package reporter

import "context"

type reporterKeyType struct{}

var reporterKeyVal reporterKeyType

func NewContextWithReporter(ctx context.Context, reporter Reporter) context.Context {
	return context.WithValue(ctx, reporterKeyVal, reporter)
}

func GetReporterFromContext(ctx context.Context) (Reporter, bool) {
	v := ctx.Value(reporterKeyVal)
	if v == nil {
		return nil, false
	}

	rep, ok := v.(Reporter)

	return rep, ok
}
