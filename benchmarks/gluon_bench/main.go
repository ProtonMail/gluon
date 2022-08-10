package main

import (
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	_ "github.com/ProtonMail/gluon/benchmarks/gluon_bench/gluon_benchmarks"
	_ "github.com/ProtonMail/gluon/benchmarks/gluon_bench/imap_benchmarks"
	_ "github.com/ProtonMail/gluon/benchmarks/gluon_bench/store_benchmarks"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.ErrorLevel)
	benchmark.RunMain()
}
