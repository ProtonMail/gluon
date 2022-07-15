#/bin/bash

CLIENTS=${CLIENTS:-"200"}
SECS=${SECS:-"60"}
IMAPTEST_BIN=${IMAPTEST_BIN:-imaptest}

${IMAPTEST_BIN} host=127.0.0.1 port=1143 \
user=user1@example.com pass=password1 \
mbox=dovecot-crlf \
no_pipelining secs=${SECS} clients=${CLIENTS}
