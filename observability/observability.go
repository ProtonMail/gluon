package observability

var imapErrorMetricType int
var messageErrorMetricType int
var otherErrorMetricType int

type Sender interface {
	AddMetrics(metrics ...map[string]any)
	AddDistinctMetrics(errType any, metrics ...map[string]any)
	AddIMAPConnectionsExceededThresholdMetric(totalOpenIMAPConnections, newIMAPConnections int)
}

func SetupMetricTypes(imapErrorType, messageErrorType, otherErrorType int) {
	imapErrorMetricType = imapErrorType
	messageErrorMetricType = messageErrorType
	otherErrorMetricType = otherErrorType
}
