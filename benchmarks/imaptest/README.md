# ImapTest  - Profiling & Compliance

This "benchmark" uses [Dovecot's ImapTest tool](https://imapwiki.org/ImapTest) to profile the performance of Gluon. The 
information present here can also be used to verify the compliance of Gluon with the IMAP protocol.

## Installation

Follow the instructions outlined in [the tools' installation page](https://imapwiki.org/ImapTest/Installation) 
to build the binary. The test mailbox is already present in this folder.

## Running ImapTest

The bare minimum required for running ImapTest is a username and a password:

```bash
imaptest host=127.0.0.1 port=1143 user=user1@example.com pass=password1
```

For convenience, we have provided a few test scripts to quickly test Gluon with the demo binary present in this 
repository.

 * **default.sh**: Runs the default ImapTest benchmarks.
 * **full.sh**: Runs every available ImapTest operation.

Both of these scripts can be customized with the following environment variables:

* **IMAPTEST_BIN**: Location of the ImapTest binary.
* **SECS**: Number of seconds for which to run ImapTest.
* **CLIENTS**: Number of concurrent clients used by ImapTest.

Example:
```bash
# Run imaptest for 60s with one conccurent client connection.
IMAPTEST_BIN=/bin/imaptest SECS=60 CLIENTS=1 ./default.sh
```

## Note about this tool
The execution of this tool is non-deterministic, this means it can't be used to compare profile runs of two different 
versions.

It should be only be used to profile and/or stress the codebase in an isolation.