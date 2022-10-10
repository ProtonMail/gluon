package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// we redefine the json statistic types here since we don't care about the extra information

type JSONBenchmarkStatistics struct {
	Total        time.Duration
	Average      time.Duration
	Fastest      time.Duration
	Slowest      time.Duration
	Median       time.Duration
	Percentile90 time.Duration
	Percentile10 time.Duration
	RMS          time.Duration
	SampleCount  int
}

type JSONBenchmarkReport struct {
	Name       string
	Runs       []*JSONBenchmarkStatistics
	Statistics *JSONBenchmarkStatistics
}

func loadReportFromFile(path string) ([]*JSONBenchmarkReport, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var reports []*JSONBenchmarkReport

	if err := json.Unmarshal(contents, &reports); err != nil {
		return nil, err
	}

	return reports, nil
}

type BenchmarkRun struct {
	fileIndex  int
	statistics *JSONBenchmarkStatistics
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage %v json_report0 json_report1... json_report N\n", os.Args[0])
		return
	}

	// load reports
	reportFiles := os.Args[1:]
	reports := make([][]*JSONBenchmarkReport, 0, len(reportFiles))

	for _, v := range reportFiles {
		report, err := loadReportFromFile(v)
		if err != nil {
			panic(fmt.Errorf("failed to load report: %w", err))
		}

		reports = append(reports, report)
	}

	benchMap := map[string][]BenchmarkRun{}

	for idx, report := range reports {
		for _, run := range report {
			b := BenchmarkRun{fileIndex: idx, statistics: run.Statistics}

			v, ok := benchMap[run.Name]
			if ok {
				v = append(v, b)
			} else {
				v = []BenchmarkRun{b}
			}

			benchMap[run.Name] = v
		}
	}

	for k, v := range benchMap {
		if len(v) == 1 {
			fmt.Printf("Benchmark %v: Only has one run\n", k)
			continue
		}

		// check if all benchmarks have the same benchmark runs
		{
			expectedRuns := v[0].statistics.SampleCount
			for i := 1; i < len(v); i++ {
				if v[i].statistics.SampleCount != expectedRuns {
					fmt.Fprintf(os.Stderr, "Benchmark %v: File '%v' does not have the expected sample count (%v)\n",
						k, reportFiles[v[i].fileIndex], expectedRuns)
					continue
				}
			}
		}

		// Check which run has the best 90th percentile
		{
			fastest := v[0].statistics.Percentile90
			fastestIndex := 0
			for i := 1; i < len(v); i++ {
				if v[i].statistics.Percentile90 < fastest {
					fastestIndex = i
					fastest = v[i].statistics.Percentile90
				}
			}

			fmt.Printf("Benchmark %v: Fastest (90th Percentile=%v) %v\n", k, fastest, reportFiles[fastestIndex])
		}
	}
}
