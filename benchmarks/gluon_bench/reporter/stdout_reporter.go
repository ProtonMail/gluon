package reporter

import (
	"fmt"
)

// StdOutReporter prints the benchmark report to os.Stdout.
type StdOutReporter struct{}

func (*StdOutReporter) ProduceReport(reports []*BenchmarkReport) error {
	for i, v := range reports {
		fmt.Printf("[%02d] Benchmark %v\n", i, v.Name)
		fmt.Printf("[%02d] %v\n", i, v.Statistics.String())

		for r, v := range v.Runs {
			fmt.Printf("[%02d] Run %02d - %v\n", i, r, v.String())
		}
	}

	return nil
}
