package flags

import "flag"

var (
	IMAPRemoteServer          = flag.String("imap-remote-server", "", "IP address and port of the remote IMAP server to run against. E.g. 127.0.0.1:1143.")
	IMAPMessageCount          = flag.Uint("imap-msg-count", 1000, "Number of messages to add to the mailbox before each benchmark")
	IMAPRandomSeqSetIntervals = flag.Bool("imap-random-seqset-intervals", false, "When set, generate random sequence intervals rather than single numbers.")
	IMAPUIDMode               = flag.Bool("imap-uid-mode", false, "When set, will run benchmarks in UID mode if available.")
	IMAPParallelClients       = flag.Uint("imap-parallel-clients", 1, "Set the number of clients to be run in parallel during the benchmark.")
)
