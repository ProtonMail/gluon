package reporter

import (
	"fmt"

	"github.com/ProtonMail/gluon/profiling"
)

// StdOutReporter prints the benchmark report to os.Stdout.
type StdOutReporter struct{}

func (*StdOutReporter) ProduceReport(reports []*BenchmarkReport) error {
	for i, v := range reports {
		fmt.Printf("[%02d] Benchmark %v\n", i, v.Name)
		fmt.Printf("[%02d] %v\n", i, v.Statistics.String())

		for r, v := range v.Runs {
			fmt.Printf("[%02d] Run %02d - Time: %v\n", i, r, v.Duration)

			for n, v := range v.CmdStatistics {
				if v.SampleCount == 0 {
					continue
				}

				fmt.Printf("[%02d] [%02d] [%v] %v\n", i, r, profiling.CmdTypeToString(n), v.String())
			}
		}
	}

	return nil
}
