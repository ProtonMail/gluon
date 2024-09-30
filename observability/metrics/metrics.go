package metrics

import "time"

const schemaName = "bridge_gluon_errors_total"
const schemaVersion = 1

func generateGluonFailureMetric(errorType string) map[string]interface{} {
	return map[string]interface{}{
		"Name":      schemaName,
		"Version":   schemaVersion,
		"Timestamp": time.Now().Unix(),
		"Data": map[string]interface{}{
			"Value": 1,
			"Labels": map[string]string{
				"errorType": errorType,
			},
		},
	}
}

func GenerateFailedParseIMAPCommandMetric() map[string]interface{} {
	return generateGluonFailureMetric("failedParseIMAPCommand")
}

func GenerateFailedToCreateMailbox() map[string]interface{} {
	return generateGluonFailureMetric("failedCreateMailbox")
}

func GenerateFailedToDeleteMailboxMetric() map[string]interface{} {
	return generateGluonFailureMetric("failedDeleteMailbox")
}

func GenerateFailedToCopyMessagesMetric() map[string]interface{} {
	return generateGluonFailureMetric("failedCopyMessages")
}

func GenerateFailedToMoveMessagesFromMailboxMetric() map[string]interface{} {
	return generateGluonFailureMetric("failedMoveMessagesFromMailbox")
}

func GenerateFailedToRemoveDeletedMessagesMetric() map[string]interface{} {
	return generateGluonFailureMetric("failedRemoveDeletedMessages")
}

func GenerateFailedToCommitDatabaseTransactionMetric() map[string]interface{} {
	return generateGluonFailureMetric("failedCommitDatabaseTransaction")
}

func GenerateFailedToInsertMessageIntoRecoveryMailbox() map[string]interface{} {
	return generateGluonFailureMetric("failedToInsertMessageIntoRecoveryMailbox")
}
