package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	imap_benchmarks2 "github.com/ProtonMail/gluon/benchmarks/gluon_bench/imap_benchmarks"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
)

var benches = []benchmark.Benchmark{
	imap_benchmarks2.NewMailboxCreate(),
	imap_benchmarks2.NewFetch(),
	imap_benchmarks2.NewCopy(),
	imap_benchmarks2.NewMove(),
	imap_benchmarks2.NewStore(),
	imap_benchmarks2.NewExpunge(),
	imap_benchmarks2.NewSearchText(),
	imap_benchmarks2.NewSearchSince(),
}

func main() {
	var benchmarkMap = make(map[string]benchmark.Benchmark)
	for _, v := range benches {
		benchmarkMap[v.Name()] = v
	}

	flag.Usage = func() {
		fmt.Printf("Usage %v [options] benchmark0 benchmark1 ... benchmarkN\n", os.Args[0])
		fmt.Printf("\nAvailable Benchmarks:\n")

		for k := range benchmarkMap {
			fmt.Printf("  * %v\n", k)
		}

		fmt.Printf("\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	var benchmarks []benchmark.Benchmark

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		return
	}

	for _, arg := range args {
		if v, ok := benchmarkMap[arg]; ok {
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

	var benchDirConfig benchmark.BenchDirConfig
	if len(*flags.BenchPath) != 0 {
		benchDirConfig = benchmark.NewFixedBenchDirConfig(*flags.BenchPath)
	} else {
		benchDirConfig = &benchmark.TmpBenchDirConfig{}
	}

	benchmarkReports := make([]*reporter.BenchmarkReport, 0, len(benchmarks))

	for _, v := range benchmarks {
		if *flags.Verbose {
			fmt.Printf("Begin IMAPBenchmark: %v\n", v.Name())
		}

		numRuns := *flags.BenchmarkRuns

		var benchmarkRuns = make([]*reporter.BenchmarkRun, 0, numRuns)

		for r := uint(0); r < numRuns; r++ {
			if *flags.Verbose {
				fmt.Printf("IMAPBenchmark Run: %v\n", r)
			}

			benchRun := measureBenchmark(benchDirConfig, r, v)
			benchmarkRuns = append(benchmarkRuns, benchRun)
		}

		benchmarkReports = append(benchmarkReports, reporter.NewBenchmarkReport(v.Name(), benchmarkRuns...))

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

func measureBenchmark(dirConfig benchmark.BenchDirConfig, iteration uint, bench benchmark.Benchmark) *reporter.BenchmarkRun {
	benchPath, err := dirConfig.Get()

	if err != nil {
		panic(fmt.Sprintf("Failed to get server directory: %v", err))
	}

	if !*flags.ReuseState {
		benchPath = filepath.Join(benchPath, fmt.Sprintf("%v-%d", bench.Name(), iteration))
	}

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

	return benchRun
}
