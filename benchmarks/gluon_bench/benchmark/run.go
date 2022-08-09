package benchmark

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"golang.org/x/exp/slices"
)

func RunMain() {
	flag.Usage = func() {
		fmt.Printf("Usage %v [options] benchmark0 benchmark1 ... benchmarkN\n", os.Args[0])
		fmt.Printf("\nAvailable Benchmarks:\n")

		var benchmarks []string

		for k := range GetBenchmarks() {
			benchmarks = append(benchmarks, k)
		}

		slices.Sort(benchmarks)

		for _, k := range benchmarks {
			fmt.Printf("  * %v\n", k)
		}

		fmt.Printf("\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	var benchmarks []Benchmark

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		return
	}

	for _, arg := range args {
		if v, ok := GetBenchmarks()[arg]; ok {
			benchmarks = append(benchmarks, v)
		}
	}

	if len(benchmarks) == 0 {
		panic("No benchmarks selected")
	}

	var benchmarkReporter reporter.BenchmarkReporter

	if len(*flags.JsonReporter) != 0 {
		benchmarkReporter = reporter.NewJSONReporter(*flags.JsonReporter)
	} else {
		benchmarkReporter = &reporter.StdOutReporter{}
	}

	var benchDirConfig BenchDirConfig
	if len(*flags.BenchPath) != 0 {
		benchDirConfig = NewFixedBenchDirConfig(*flags.BenchPath)
	} else {
		benchDirConfig = &TmpBenchDirConfig{}
	}

	benchmarkReports := make([]*reporter.BenchmarkReport, 0, len(benchmarks))

	for _, v := range benchmarks {
		if *flags.Verbose {
			fmt.Printf("Begin IMAPBenchmark: %v\n", v.Name())
		}

		numRuns := *flags.BenchmarkRuns

		var benchmarkStats = make([]*reporter.BenchmarkStatistics, 0, numRuns)

		for r := uint(0); r < numRuns; r++ {
			if *flags.Verbose {
				fmt.Printf("IMAPBenchmark Run: %v\n", r)
			}

			benchStat := measureBenchmark(benchDirConfig, r, v)
			benchmarkStats = append(benchmarkStats, benchStat)
		}

		benchmarkReports = append(benchmarkReports, reporter.NewBenchmarkReport(v.Name(), benchmarkStats...))

		if *flags.Verbose {
			fmt.Printf("End IMAPBenchmark: %v\n", v.Name())
		}
	}

	if benchmarkReporter != nil {
		if *flags.Verbose {
			fmt.Printf("Generating Report\n")
		}

		if err := benchmarkReporter.ProduceReport(benchmarkReports); err != nil {
			panic(fmt.Sprintf("Failed to produce benchmark report: %v", err))
		}
	}

	if *flags.Verbose {
		fmt.Printf("Finished\n")
	}
}

func measureBenchmark(dirConfig BenchDirConfig, iteration uint, bench Benchmark) *reporter.BenchmarkStatistics {
	benchPath, err := dirConfig.Get()

	if err != nil {
		panic(fmt.Sprintf("Failed to get server directory: %v", err))
	}

	benchPath = filepath.Join(benchPath, fmt.Sprintf("%v-%d", bench.Name(), iteration))

	if *flags.Verbose {
		fmt.Printf("IMAPBenchmark Data Path: %v\n", benchPath)
	}

	if err := os.MkdirAll(benchPath, 0o777); err != nil {
		panic(fmt.Sprintf("Failed to create server directory '%v' : %v", benchPath, err))
	}

	ctx := context.Background()
	if err := bench.Setup(ctx, benchPath); err != nil {
		panic(fmt.Sprintf("Failed to setup benchmark %v: %v", bench.Name(), err))
	}

	benchRun, benchErr := bench.Run(ctx)
	if benchErr != nil {
		panic(fmt.Sprintf("Failed to run benchmark %v: %v", bench.Name(), err))
	}

	if err := bench.TearDown(ctx); err != nil {
		panic(fmt.Sprintf("Failed to teardown benchmark %v: %v", bench.Name(), err))
	}

	return reporter.NewBenchmarkStatistics(benchRun.Extra, benchRun.Durations...)
}
