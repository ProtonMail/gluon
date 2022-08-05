package flags

import "flag"

var (
	RemoteServer          = flag.String("remote-server", "", "IP address and port of the remote IMAP server to run against. E.g. 127.0.0.1:1143.")
	MessageCount          = flag.Uint("msg-count", 1000, "Number of messages to add to the mailbox before each benchmark")
	RandomSeqSetIntervals = flag.Bool("random-seqset-intervals", false, "When set, generate random sequence intervals rather than single numbers.")
	UIDMode               = flag.Bool("uid-mode", false, "When set, will run benchmarks in UID mode if available.")
	ParallelClients       = flag.Uint("parallel-clients", 1, "Set the number of clients to be run in parallel during the benchmark.")
)
