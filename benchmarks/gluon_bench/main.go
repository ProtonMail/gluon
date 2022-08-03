package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmarks"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/reporter"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/server"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
	"github.com/ProtonMail/gluon/profiling"
)

var benches = []benchmarks.Benchmark{
	&benchmarks.MailboxCreate{},
	benchmarks.NewFetch(),
	benchmarks.NewCopy(),
	benchmarks.NewMove(),
	benchmarks.NewStore(),
	benchmarks.NewExpunge(),
	benchmarks.NewSearchText(),
	benchmarks.NewSearchSince(),
}

func main() {
	var benchmarkMap = make(map[string]benchmarks.Benchmark)
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

	if len(*flags.JsonReporter) != 0 {
		benchmarkReporter = reporter.NewJSONReporter(*flags.JsonReporter)
	} else {
		benchmarkReporter = &reporter.StdOutReporter{}
	}

	cmdProfiler := utils.NewDurationCmdProfilerBuilder()

	var serverDirConfig utils.ServerDirConfig
	if len(*flags.StorePath) != 0 {
		serverDirConfig = utils.NewPersistentServerDirConfig(*flags.StorePath)
	} else {
		serverDirConfig = &utils.TmpServerDirConfig{}
	}

	var serverBuilder server.ServerBuilder

	if len(*flags.RemoteServer) != 0 {
		builder, err := server.NewRemoteServerBuilder(*flags.RemoteServer)
		if err != nil {
			panic(fmt.Sprintf("Invalid Server address: %v", err))
		}

		serverBuilder = builder
	} else {
		serverBuilder = &server.LocalServerBuilder{}
	}

	benchmarkReports := make([]*reporter.BenchmarkReport, 0, len(benchmarks))

	for _, v := range benchmarks {
		if *flags.Verbose {
			fmt.Printf("Begin Benchmark: %v\n", v.Name())
		}

		numRuns := *flags.BenchmarkRuns

		var benchmarkRuns = make([]*reporter.BenchmarkRun, 0, numRuns)

		for r := uint(0); r < numRuns; r++ {
			if *flags.Verbose {
				fmt.Printf("Benchmark Run: %v\n", r)
			}

			cmdProfiler.Clear()

			scopedTimer := measureBenchmark(serverDirConfig, serverBuilder, r, cmdProfiler, v)
			benchmarkRuns = append(benchmarkRuns, reporter.NewBenchmarkRun(scopedTimer.Elapsed(), cmdProfiler.Merge()))
		}

		benchmarkReports = append(benchmarkReports, reporter.NewBenchmarkReport(v.Name(), benchmarkRuns...))

		if *flags.Verbose {
			fmt.Printf("End Benchmark: %v\n", v.Name())
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

func measureBenchmark(dirConfig utils.ServerDirConfig, serverBuilder server.ServerBuilder, iteration uint,
	cmdProfiler profiling.CmdProfilerBuilder, bench benchmarks.Benchmark) utils.ScopedTimer {
	serverPath, err := dirConfig.Get()

	if err != nil {
		panic(fmt.Sprintf("Failed to get server directory: %v", err))
	}

	if !*flags.ReuseState {
		serverPath = filepath.Join(serverPath, fmt.Sprintf("%v-%d", bench.Name(), iteration))
	}

	if *flags.Verbose {
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
