#/bin/bash

CLIENTS=${CLIENTS:-"200"}
SECS=${SECS:-"60"}
IMAPTEST_BIN=${IMAPTEST_BIN:-imaptest}

$IMAPTEST_BIN host=127.0.0.1 port=1143 \
user=user1@example.com pass=password1 mbox=dovecot-crlf \
no_pipelining secs=${SECS} clients=${CLIENTS} \
- mcreate=50 \
mdelete=50 \
uidf=50 \
search=30 \
noop=15 \
fetch=50 \
login=100 \
logout=100 \
list=50 \
select=100 \
fet2=100,30 \
copy=30,5 \
store=50 \
delete=100 \
expunge=100 \
append=100,5
