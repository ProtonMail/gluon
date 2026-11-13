package reporter

type Context = map[string]any
type Tags = map[string]string

// Reporter represents an external reporting tool which can be hooked into gluon to report key information and/or
// unexpected behaviors.
type Reporter interface {
	ReportException(any) error
	ReportMessage(string) error
	ReportMessageWithContext(string, Context) error
	ReportWarningWithContext(string, Context) error
	ReportMessageWithContextAndTags(string, Context, Tags) error
	ReportExceptionWithContext(any, Context) error
}
