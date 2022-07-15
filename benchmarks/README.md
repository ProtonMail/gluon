# Gluon Benchmarks

Collection of benchmarks and/or tools to test and/or profile Gluon performance. See each folder's `README.md` for more information.


## Capturing Profiles

You can use the gluon demo binary as a way to profile Gluon. Be sure to build the binary with `go build`.

You need to run the benchmark for each type of profile you wish to generate as it is not possible to profile
multiple items at the same time.

Example:
```bash
./demo -profile-cpu -profile-path=$PWD # Generates $PWD/cpu.pprof - CPU Profiling.
./demo -profile-mem -profile-path=$PWD # Generates $PWD/mem.pprof - Memory Profiling.
./demo -profile-lock -profile-path=$PWD # Generates $PWD/block.pprof - Locking/blocking profiling.
```

When the benchmark has finished execution kill the demo binary and the profile data will be written out.

## Analysing the data

You need to install the pprof  tool (`go install github.com/google/pprof`) to analyse the generated data. 

After the tool is installed you can analyse the data with `pprof <path to .pprof file>`. See [this blog post](1) and 
[this tutorial](2) for an introduction to pprof.

For a more comprehensible overview you can also launch pprof's web interface which includes many additions
such as a flamegraph.

```bash
pprof -http 127.0.0.1:8080 <path to .pprof file>
```

## Available Benchmarks/Tools
* [imaptest](imaptest)

## References
 
[1]: https://www.youtube.com/watch?v=N3PWzBeLX2M

[2]: https://go.dev/blog/pprof