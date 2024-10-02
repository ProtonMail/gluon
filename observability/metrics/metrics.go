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

// TODO (atanas) maybe think about a different root table here
func GenerateAppendToDraftsMustNotReturnExistingRemoteID() map[string]interface{} {
	return generateGluonFailureMetric("appendToDraftsReturnedExistingRemoteID")
}

func GenerateDatabaseMigrationFailed() map[string]interface{} {
	return generateGluonFailureMetric("databaseMigrationFailed")
}

func GenerateAllMetrics() []map[string]interface{} {
	var metrics []map[string]interface{}
	metrics = append(metrics,
		GenerateFailedParseIMAPCommandMetric(),
		GenerateFailedToCreateMailbox(),
		GenerateFailedToDeleteMailboxMetric(),
		GenerateFailedToCopyMessagesMetric(),
		GenerateFailedToMoveMessagesFromMailboxMetric(),
		GenerateFailedToRemoveDeletedMessagesMetric(),
		GenerateFailedToCommitDatabaseTransactionMetric(),
		GenerateAppendToDraftsMustNotReturnExistingRemoteID(),
		GenerateDatabaseMigrationFailed(),
	)

	return metrics
}
