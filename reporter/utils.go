package reporter

import (
	"context"

	"github.com/sirupsen/logrus"
)

func MessageWithContext(ctx context.Context, message string, context Context) {
	reporter, ok := GetReporterFromContext(ctx)
	if !ok {
		return
	}

	if err := reporter.ReportMessageWithContext(message, context); err != nil {
		logrus.WithError(err).Error("Failed to report message")
	}
}

func ExceptionWithContext(ctx context.Context, message string, context Context) {
	reporter, ok := GetReporterFromContext(ctx)
	if !ok {
		return
	}

	if err := reporter.ReportExceptionWithContext(message, context); err != nil {
		logrus.WithError(err).Error("Failed to report message")
	}
}

func Exception(ctx context.Context, info any) {
	reporter, ok := GetReporterFromContext(ctx)
	if !ok {
		return
	}

	if err := reporter.ReportException(info); err != nil {
		logrus.WithError(err).Error("Failed to report message")
	}
}

func Message(ctx context.Context, message string) {
	reporter, ok := GetReporterFromContext(ctx)
	if !ok {
		return
	}

	if err := reporter.ReportMessage(message); err != nil {
		logrus.WithError(err).Error("Failed to report message")
	}
}
