package flags

import "flag"

var (
	BenchPath     = flag.String("path", "", "Filepath where to write the database data. If not set a temp folder will be used.")
	Verbose       = flag.Bool("verbose", false, "Enable verbose logging.")
	JsonReporter  = flag.String("json-reporter", "", "If specified, will generate a json report with the given filename.")
	BenchmarkRuns = flag.Uint("bench-runs", 1, "Number of runs per benchmark.")
	Connector     = flag.String("connector", "dummy", "Key of the connector implementation registered with ConnectorFactory.")
	UserName      = flag.String("user-name", "user", "Username for the connector user, defaults to 'user'.")
	UserPassword  = flag.String("user-pwd", "password", "Password for the connector user, defaults to 'password'.")
	SkipClean     = flag.Bool("skip-clean", false, "Do not cleanup benchmark data directory.")
)
