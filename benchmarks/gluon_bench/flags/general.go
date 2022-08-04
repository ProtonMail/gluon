package flags

import "flag"

var (
	BenchPath     = flag.String("path", "", "Filepath where to write the database data. If not set a temp folder will be used.")
	Verbose       = flag.Bool("verbose", false, "Enable verbose logging.")
	JsonReporter  = flag.String("json-reporter", "", "If specified, will generate a json report with the given filename.")
	BenchmarkRuns = flag.Uint("bench-runs", 1, "Number of runs per benchmark.")
	ReuseState    = flag.Bool("reuse-state", false, "When present, benchmarks will re-use previous run state, rather than a clean slate.")
)
