package reporter

type NullReporter struct{}

func (*NullReporter) ReportException(any) error {
	return nil
}

func (*NullReporter) ReportMessage(string) error {
	return nil
}

func (*NullReporter) ReportMessageWithContext(string, Context) error {
	return nil
}

func (*NullReporter) ReportExceptionWithContext(any, Context) error {
	return nil
}
