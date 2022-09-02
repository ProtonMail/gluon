package reporter

type Context = map[string]any

// Reporter represents an external reporting tool which can be hooked into gluon to report key information and/or
// unexpected behaviors.
type Reporter interface {
	ReportException(any) error
	ReportMessage(string) error
	ReportMessageWithContext(string, Context) error
	ReportExceptionWithContext(any, Context) error
}
