package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmarks"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/server"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
	"github.com/ProtonMail/gluon/profiling"
)

var storePathFlag = flag.String("path", "", "Filepath where to write the database data. If not set a temp folder will be used.")
var verboseFlag = flag.Bool("verbose", false, "Enable verbose logging.")
var jsonReporterFlag = flag.String("json-reporter", "", "If specified, will generate a json report with the given filename.")
var benchmarkRunsFlag = flag.Uint("bench-runs", 1, "Number of runs per benchmark.")
var reuseStateFlag = flag.Bool("reuse-state", false, "When present, benchmarks will re-use previous run state, rather than a clean slate.")
var remoteServerFlag = flag.String("remote-server", "", "IP address and port of the remote IMAP server to run against. E.g. 127.0.0.1:1143.")

var benches = []benchmarks.Benchmark{
	&benchmarks.MailboxCreate{},
}

func main() {
	var benchmarkMap = make(map[string]benchmarks.Benchmark)
	for _, v := range benches {
		benchmarkMap[v.Name()] = v
	}

	flag.Usage = func() {
		fmt.Printf("Usage %v [options] benchmark0 benchmark1 ... benchmarkN\n", os.Args[0])
		fmt.Printf("\nAvailable Benchmarks:\n")

		for k, _ := range benchmarkMap {
			fmt.Printf("  * %v\n", k)
		}

		fmt.Printf("\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	var benchmarks []benchmarks.Benchmark

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

	if len(*jsonReporterFlag) != 0 {
		benchmarkReporter = reporter.NewJSONReporter(*jsonReporterFlag)
	} else {
		benchmarkReporter = &reporter.StdOutReporter{}
	}

	cmdProfiler := utils.NewDurationCmdProfilerBuilder()

	var serverDirConfig utils.ServerDirConfig
	if len(*storePathFlag) != 0 {
		serverDirConfig = utils.NewPersistentServerDirConfig(*storePathFlag)
	} else {
		serverDirConfig = &utils.TmpServerDirConfig{}
	}

	var serverBuilder server.ServerBuilder

	if len(*remoteServerFlag) != 0 {
		builder, err := server.NewRemoteServerBuilder(*remoteServerFlag)
		if err != nil {
			panic(fmt.Sprintf("Invalid Server address: %v", err))
		}

		serverBuilder = builder
	} else {
		serverBuilder = &server.LocalServerBuilder{}
	}

	benchmarkReports := make([]*reporter.BenchmarkReport, 0, len(benchmarks))

	for _, v := range benchmarks {
		if *verboseFlag {
			fmt.Printf("Begin Benchmark: %v\n", v.Name())
		}

		numRuns := *benchmarkRunsFlag

		var benchmarkRuns = make([]*reporter.BenchmarkRun, 0, numRuns)

		for r := uint(0); r < numRuns; r++ {
			if *verboseFlag {
				fmt.Printf("Benchmark Run: %v\n", r)
			}

			cmdProfiler.Clear()

			scopedTimer := measureBenchmark(serverDirConfig, serverBuilder, r, cmdProfiler, v)
			benchmarkRuns = append(benchmarkRuns, reporter.NewBenchmarkRun(scopedTimer.Elapsed(), cmdProfiler.Merge()))
		}

		benchmarkReports = append(benchmarkReports, reporter.NewBenchmarkReport(v.Name(), benchmarkRuns...))

		if *verboseFlag {
			fmt.Printf("End Benchmark: %v\n", v.Name())
		}
	}

	if benchmarkReporter != nil {
		if *verboseFlag {
			fmt.Printf("Generating Report\n")
		}

		if err := benchmarkReporter.ProduceReport(benchmarkReports); err != nil {
			panic(fmt.Sprintf("Failed to produce benchmark report: %v", err))
		}
	}

	if *verboseFlag {
		fmt.Printf("Finished\n")
	}
}

func measureBenchmark(dirConfig utils.ServerDirConfig, serverBuilder server.ServerBuilder, iteration uint,
	cmdProfiler profiling.CmdProfilerBuilder, bench benchmarks.Benchmark) utils.ScopedTimer {
	serverPath, err := dirConfig.Get()

	if err != nil {
		panic(fmt.Sprintf("Failed to get server directory: %v", err))
	}

	if !*reuseStateFlag {
		serverPath = filepath.Join(serverPath, fmt.Sprintf("%v-%d", bench.Name(), iteration))
	} else {
		serverPath = filepath.Join(serverPath, bench.Name())
	}

	if *verboseFlag {
		fmt.Printf("Benchmark Data Path: %v\n", serverPath)
	}

	if err := os.MkdirAll(serverPath, 0o777); err != nil {
		panic(fmt.Sprintf("Failed to create server directory '%v' : %v", serverPath, err))
	}

	ctx := context.Background()
	server, err := serverBuilder.New(ctx, serverPath, cmdProfiler)

	if err != nil {
		panic(fmt.Sprintf("Failed to create server: %v", err))
	}

	address := server.Address()

	if err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}

	defer server.Close(ctx)

	if err := bench.Setup(ctx, address); err != nil {
		panic(fmt.Sprintf("Failed to setup benchmark %v: %v", bench.Name(), err))
	}

	scopedTimer := utils.ScopedTimer{}
	scopedTimer.Start()

	benchErr := bench.Run(ctx, address)

	scopedTimer.Stop()

	if benchErr != nil {
		panic(fmt.Sprintf("Failed to run benchmark %v: %v", bench.Name(), err))
	}

	if err := bench.TearDown(ctx, address); err != nil {
		panic(fmt.Sprintf("Failed to teardown benchmark %v: %v", bench.Name(), err))
	}

	return scopedTimer
}
