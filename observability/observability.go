package observability

var imapErrorMetricType int
var messageErrorMetricType int
var otherErrorMetricType int

type Sender interface {
	AddMetrics(metrics ...map[string]interface{})
	AddDistinctMetrics(errType interface{}, metrics ...map[string]interface{})
	AddIMAPConnectionsExceededThresholdMetric(totalOpenIMAPConnections, newIMAPConnections int)
}

func SetupMetricTypes(imapErrorType, messageErrorType, otherErrorType int) {
	imapErrorMetricType = imapErrorType
	messageErrorMetricType = messageErrorType
	otherErrorMetricType = otherErrorType
}
