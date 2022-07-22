package flags

import "flag"

var StorePath = flag.String("path", "", "Filepath where to write the database data. If not set a temp folder will be used.")
var Verbose = flag.Bool("verbose", false, "Enable verbose logging.")
var JsonReporter = flag.String("json-reporter", "", "If specified, will generate a json report with the given filename.")
var BenchmarkRuns = flag.Uint("bench-runs", 1, "Number of runs per benchmark.")
var ReuseState = flag.Bool("reuse-state", false, "When present, benchmarks will re-use previous run state, rather than a clean slate.")
var RemoteServer = flag.String("remote-server", "", "IP address and port of the remote IMAP server to run against. E.g. 127.0.0.1:1143.")
var Mailbox = flag.String("mailbox", "INBOX", "If not specified will use INBOX as the mailbox to run benchmarks against.")
var ParallelClients = flag.Uint("parallel-clients", 1, "Set the number of clients to be run in parallel during the benchmark.")
var FillSourceMailbox = flag.Uint("fill-src-mailbox", 1000, "Number of messages to add to the source inbox before each benchmark, set to 0 to skip.")
var RandomSeqSetIntervals = flag.Bool("random-seqset-intervals", false, "When set, generate random sequence intervals rather than single numbers.")
var UIDMode = flag.Bool("uid-mode", false, "When set, will run benchmarks in UID mode if available.")
