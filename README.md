<p align="center">
<h1 align="center">Gluon</h1>
<p align="center">An <a href="https://datatracker.ietf.org/doc/html/rfc3501">IMAP4rev1</a> library focusing on correctness, stability and performance.</p>
<p align="center">
<a href="https://github.com/ProtonMail/gluon/actions/workflows/release.yml"><img src="https://github.com/ProtonMail/gluon/actions/workflows/release.yml/badge.svg?branch=master" alt="CI Status"></a>
<a href="https://pkg.go.dev/github.com/ProtonMail/gluon"><img src="https://pkg.go.dev/badge/github.com/ProtonMail/gluon" alt="GoDoc"></a>
<a href="https://goreportcard.com/report/github.com/ProtonMail/gluon"><img src="https://goreportcard.com/badge/github.com/ProtonMail/gluon" alt="Go Report Card"></a>
<a href="LICENSE"><img src="https://img.shields.io/github/license/ProtonMail/gluon.svg" alt="License"></a>
</p>

# Demo

The demo server can be started with:

```
$ GLUON_LOG_LEVEL=trace go run demo/demo.go
DEBU[0000] Applying update                               update="MailboxCreated: Mailbox.ID = 0, Mailbox.Name = INBOX" user-id=ac8970c5-cdb7-4043-ad85-ad9b9defcfb8
DEBU[0000] Applying update                               update="MessagesCreated: MessageCount=0 Messages=[]" user-id=ac8970c5-cdb7-4043-ad85-ad9b9defcfb8
INFO[0000] User added to server                          userID=ac8970c5-cdb7-4043-ad85-ad9b9defcfb8
DEBU[0000] Applying update                               update="MailboxCreated: Mailbox.ID = 0, Mailbox.Name = INBOX" user-id=a51fad46-9bde-462a-a467-6c30f9a40a63
DEBU[0000] Applying update                               update="MessagesCreated: MessageCount=0 Messages=[]" user-id=a51fad46-9bde-462a-a467-6c30f9a40a63
INFO[0000] User added to server                          userID=a51fad46-9bde-462a-a467-6c30f9a40a63
INFO[0000] Server is listening on 127.0.0.1:1143
```

By default, the demo server includes two demo users, both with password `pass`.
The first has addresses `user1@example.com` and `alias1@example.com`.
The second has addresses `user2@example.com` and `alias2@example.com`.

Once started, connect to the demo server with an email client (e.g. thunderbird) or via telnet:
```
$ telnet 127.0.0.1 1143
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
* OK [CAPABILITY IDLE IMAP4rev1 MOVE UIDPLUS UNSELECT]  00.00.00 - gluon session ID 2
tag login user1@example.com pass
tag OK [CAPABILITY IDLE IMAP4rev1 MOVE UIDPLUS UNSELECT] Logged in
tag append inbox (\Seen) {14}
+ Ready
To: user@pm.me
tag OK [APPENDUID 1 1] APPEND
tag select inbox
* FLAGS (\Answered \Deleted \Flagged \Seen)
* 1 EXISTS
* 1 RECENT
* OK [PERMANENTFLAGS (\Answered \Deleted \Flagged \Seen)] Flags permitted
* OK [UIDNEXT 2] Predicted next UID
* OK [UIDVALIDITY 1] UIDs valid
tag OK [READ-WRITE] SELECT
tag fetch 1:* (UID BODY.PEEK[])
* 1 FETCH (UID 1 BODY[] {32}
X-Pm-Gluon-Id: 1
To: user@pm.me)
tag OK command completed in 1.030958ms
```

# Changing DB schema

Do not forget to re-generate ent code after changing the DB schema in `./internal/db/ent/schema`.

```
pushd ./internal/db/ent && go generate . && popd

