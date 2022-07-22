package flags

import "flag"

var StorePathFlag = flag.String("path", "", "Filepath where to write the database data. If not set a temp folder will be used.")
var VerboseFlag = flag.Bool("verbose", false, "Enable verbose logging.")
var JsonReporterFlag = flag.String("json-reporter", "", "If specified, will generate a json report with the given filename.")
var BenchmarkRunsFlag = flag.Uint("bench-runs", 1, "Number of runs per benchmark.")
var ReuseStateFlag = flag.Bool("reuse-state", false, "When present, benchmarks will re-use previous run state, rather than a clean slate.")
var RemoteServerFlag = flag.String("remote-server", "", "IP address and port of the remote IMAP server to run against. E.g. 127.0.0.1:1143.")
var MailboxFlag = flag.String("mailbox", "INBOX", "If not specified will use INBOX as the mailbox to run benchmarks against.")
var ParallelClientsFlag = flag.Uint("parallel-clients", 1, "Set the number of clients to be run in parallel during the benchmark.")
var FillSourceMailbox = flag.Uint("fill-src-mailbox", 1000, "Number of messages to add to the source inbox before each benchmark, set to 0 to skip.")
var FlagRandomSeqSetIntervals = flag.Bool("random-seqset-intervals", false, "When set, generate random sequence intervals rather than single numbers.")
