# ImapTest  - Profiling & Compliance

This "benchmark" uses [Dovecot's ImapTest tool](https://imapwiki.org/ImapTest) to profile the performance of Gluon. The 
information present here can also be used to verify the compliance of Gluon with the IMAP protocol.

Build gluon demo in project root folder:

```bash
go build -o gluon-demo ./demo/demo.go
```

## Installation of ImapTest

Follow the instructions outlined in [the tools' installation page](https://imapwiki.org/ImapTest/Installation) 
to build the binary. The test mailbox is already present in this folder.

## Simple ImapTest run

Assuming gluon demo is running, the bare minimum required for running ImapTest is a username and a password:

```bash
imaptest host=127.0.0.1 port=1143 user=user1@example.com pass=password1
```

## Advance testing

The multiple scenario coverage can be run by

```
go test
```

The test cases are defined in `benchmark.yml`. Each case defines a number of
clients and users to be used by ImapTest. One case can have multiple
settings defined by name. The options are specified in `settings` section. The
settings reflects ImapTest options as described in
[here](https://imapwiki.org/ImapTest/Running).
The ImapTest states are described in
[here](https://imapwiki.org/ImapTest/States).

We don't use for now the ImapTest scriptable scenarios but it is
possible by defining the new test settings in file `./benchmark.yml` and
creating separate definition file like example
[here](https://github.com/dovecot/imaptest/tree/main/src/tests)


## Note about this tool

The execution of this tool is non-deterministic, this means it can't be used to
compare profile runs of two different versions.

It should be only be used to profile and/or stress the codebase in an
isolation.
