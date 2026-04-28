package metrics

import "time"

const schemaName = "bridge_gluon_errors_total"
const schemaVersion = 1

func generateGluonErrorMetric(errorType string) map[string]any {
	return map[string]any{
		"Name":      schemaName,
		"Version":   schemaVersion,
		"Timestamp": time.Now().Unix(),
		"Data": map[string]any{
			"Value": 1,
			"Labels": map[string]string{
				"errorType": errorType,
			},
		},
	}
}

func GenerateFailedParseIMAPCommandMetric() map[string]any {
	return generateGluonErrorMetric("failedParseIMAPCommand")
}

func GenerateFailedToCreateMailbox() map[string]any {
	return generateGluonErrorMetric("failedCreateMailbox")
}

func GenerateFailedToDeleteMailboxMetric() map[string]any {
	return generateGluonErrorMetric("failedDeleteMailbox")
}

func GenerateFailedToCopyMessagesMetric() map[string]any {
	return generateGluonErrorMetric("failedCopyMessages")
}

func GenerateFailedToMoveMessagesFromMailboxMetric() map[string]any {
	return generateGluonErrorMetric("failedMoveMessagesFromMailbox")
}

func GenerateFailedToRemoveDeletedMessagesMetric() map[string]any {
	return generateGluonErrorMetric("failedRemoveDeletedMessages")
}

func GenerateFailedToCommitDatabaseTransactionMetric() map[string]any {
	return generateGluonErrorMetric("failedCommitDatabaseTransaction")
}

func GenerateAppendToDraftsMustNotReturnExistingRemoteID() map[string]any {
	return generateGluonErrorMetric("appendToDraftsReturnedExistingRemoteID")
}

func GenerateDatabaseMigrationFailed() map[string]any {
	return generateGluonErrorMetric("databaseMigrationFailed")
}

func GenerateFailedToStoreFlagsOnMessages() map[string]any {
	return generateGluonErrorMetric("failedToStoreFlagsOnMessages")
}
