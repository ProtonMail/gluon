# Gluon Bench - IMAP benchmarks

Gluon bench provides a collection of benchmarks that operate either at the IMAP client level or directly on Gluon 
itself (e.g: sync).

All IMAP command related benchmarks can be run against a local gluon server which will be started with the benchmark or 
an externally running IMAP server.

If running against a local server, it's possible to record the execution times of every individual command. 

Finally, it is also possible to produce a JSON report rather than printing to the console.


## Building

```bash
# In benchmarks/gluon_bench
go build main.go -o gluon_bench 
```

## Running Gluon Bench

To run Gluon Bench specify a set of options followed by a set of benchmarks you wish to run:

```bash
gluon_bench -verbose -parallel-client=4 fetch append
```

Please consult the output of `gluon_bench  -h` for all available options/modifiers and benchmarks.


## Integrating Gluon Bench in other projects

When integrating Gluon Bench in other projects which may contain other gluon connectors:

* Register your connector with `utils.RegisterConnector()`
* Specify the connector with the option `-connector=<...>`
* In your `main` call `benchmark.RunMain()`