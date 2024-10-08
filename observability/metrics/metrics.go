package metrics

import "time"

const schemaName = "bridge_gluon_errors_total"
const schemaVersion = 1

func generateGluonErrorMetric(errorType string) map[string]interface{} {
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
	return generateGluonErrorMetric("failedParseIMAPCommand")
}

func GenerateFailedToCreateMailbox() map[string]interface{} {
	return generateGluonErrorMetric("failedCreateMailbox")
}

func GenerateFailedToDeleteMailboxMetric() map[string]interface{} {
	return generateGluonErrorMetric("failedDeleteMailbox")
}

func GenerateFailedToCopyMessagesMetric() map[string]interface{} {
	return generateGluonErrorMetric("failedCopyMessages")
}

func GenerateFailedToMoveMessagesFromMailboxMetric() map[string]interface{} {
	return generateGluonErrorMetric("failedMoveMessagesFromMailbox")
}

func GenerateFailedToRemoveDeletedMessagesMetric() map[string]interface{} {
	return generateGluonErrorMetric("failedRemoveDeletedMessages")
}

func GenerateFailedToCommitDatabaseTransactionMetric() map[string]interface{} {
	return generateGluonErrorMetric("failedCommitDatabaseTransaction")
}

func GenerateAppendToDraftsMustNotReturnExistingRemoteID() map[string]interface{} {
	return generateGluonErrorMetric("appendToDraftsReturnedExistingRemoteID")
}

func GenerateDatabaseMigrationFailed() map[string]interface{} {
	return generateGluonErrorMetric("databaseMigrationFailed")
}
