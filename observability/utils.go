package observability

import "context"

func AddImapMetric(ctx context.Context, metric ...map[string]interface{}) {
	sender, ok := getObservabilitySenderFromContext(ctx)
	if !ok {
		return
	}

	sender.AddDistinctMetrics(imapErrorMetricType, metric...)
}

func AddMessageRelatedMetric(ctx context.Context, metric ...map[string]interface{}) {
	sender, ok := getObservabilitySenderFromContext(ctx)
	if !ok {
		return
	}

	sender.AddDistinctMetrics(messageErrorMetricType, metric...)
}

func AddOtherMetric(ctx context.Context, metric ...map[string]interface{}) {
	sender, ok := getObservabilitySenderFromContext(ctx)
	if !ok {
		return
	}

	sender.AddDistinctMetrics(otherErrorMetricType, metric...)
}
