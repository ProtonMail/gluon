package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmarks"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
	"github.com/ProtonMail/gluon/profiling"
)

var storePathFlag = flag.String("path", "", "Filepath where to write the database data. If not set a temp folder will be used.")
var verboseFlag = flag.Bool("verbose", false, "Enable verbose logging.")
var jsonReporterFlag = flag.String("json-reporter", "", "If specified, will generate a json report with the given filename.")
var benchmarkRunsFlag = flag.Uint("bench-runs", 1, "Number of runs per benchmark.")
var reuseStateFlag = flag.Bool("reuse-state", false, "When present, benchmarks will re-use previous run state, rather than a clean slate.")

var benchmarkMap = map[string]func(context.Context, *gluon.Server, string){
	"bench-mailbox-create": benchmarks.BenchmarkMailboxCreate,
}

type benchmarkEntry struct {
	key     string
	benchFn func(context.Context, *gluon.Server, string)
}

func main() {
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

	var benchmarks []benchmarkEntry

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		return
	}

	for _, arg := range args {
		if v, ok := benchmarkMap[arg]; ok {
			benchmarks = append(benchmarks, benchmarkEntry{key: arg, benchFn: v})
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

	benchmarkReports := make([]*reporter.BenchmarkReport, 0, len(benchmarks))

	for _, v := range benchmarks {
		if *verboseFlag {
			fmt.Printf("Begin Benchmark: %v\n", v.key)
		}

		numRuns := *benchmarkRunsFlag

		var benchmarkRuns = make([]*reporter.BenchmarkRun, 0, numRuns)

		for r := uint(0); r < numRuns; r++ {
			if *verboseFlag {
				fmt.Printf("Benchmark Run: %v\n", r)
			}

			cmdProfiler.Clear()

			scopedTimer := measureBenchmark(serverDirConfig, v.key, r, cmdProfiler, v.benchFn)
			benchmarkRuns = append(benchmarkRuns, reporter.NewBenchmarkRun(scopedTimer.Elapsed(), cmdProfiler.Merge()))
		}

		benchmarkReports = append(benchmarkReports, reporter.NewBenchmarkReport(v.key, benchmarkRuns...))

		if *verboseFlag {
			fmt.Printf("End Benchmark: %v\n", v.key)
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

func measureBenchmark(dirConfig utils.ServerDirConfig, name string, iteration uint, cmdProfiler profiling.CmdProfilerBuilder, bench func(context.Context, *gluon.Server, string)) utils.ScopedTimer {
	serverPath, err := dirConfig.Get()
	if err != nil {
		panic(fmt.Sprintf("Failed to get server directory: %v", err))
	}

	if !*reuseStateFlag {
		serverPath = filepath.Join(serverPath, fmt.Sprintf("%v-%d", name, iteration))
	} else {
		serverPath = filepath.Join(serverPath, name)
	}

	if *verboseFlag {
		fmt.Printf("Benchmark Data Path: %v\n", serverPath)
	}

	if err := os.MkdirAll(serverPath, 0o777); err != nil {
		panic(fmt.Sprintf("Failed to create server directory '%v' : %v", serverPath, err))
	}

	ctx := context.Background()
	server, address, err := newServer(ctx, serverPath, gluon.WithCmdProfiler(cmdProfiler))

	if err != nil {
		panic(fmt.Sprintf("Failed to start server: %v", err))
	}

	defer server.Close(ctx)

	scopedTimer := utils.ScopedTimer{}
	scopedTimer.Start()
	bench(ctx, server, address)
	scopedTimer.Stop()

	return scopedTimer
}
