<p align="center">
<h1 align="center">Gluon</h1>
<p align="center">An <a href="https://datatracker.ietf.org/doc/html/rfc3501">IMAP4rev1</a> library focusing on correctness, stability and performance.</p>
<p align="center">This is work in progress. It will eventually be integrated into the Proton Bridge.</p>
<p align="center">
<a href="https://github.com/ProtonMail/gluon/actions/workflows/release.yml"><img src="https://github.com/ProtonMail/gluon/actions/workflows/release.yml/badge.svg?branch=master" alt="CI Status"></a>
<a href="https://pkg.go.dev/github.com/ProtonMail/gluon"><img src="https://pkg.go.dev/badge/github.com/ProtonMail/gluon" alt="GoDoc"></a>
<a href="https://goreportcard.com/report/ProtonMail/gluon"><img src="https://goreportcard.com/badge/ProtonMail/gluon" alt="Go Report Card"></a>
<a href="LICENSE"><img src="https://img.shields.io/github/license/ProtonMail/gluon.svg" alt="License"></a>
</p>

# Demo

The demo server can be started with:

```
$ go run demo/demo.go
INFO[0000] User added to server                          userID=d5a706ae-c7bf-4cfd-bad3-982eafcdfe39
INFO[0000] User added to server                          userID=622e121e-c9c7-43f9-b0ee-22bf868e8429
INFO[0000] Server is listening on 127.0.0.1:1143
```

The demo server includes two demo users,
the first has addresses `user1@example.com` and `alias1@example.com` and password `password1`,
the second has addresses `user2@example.com` and `alias2@example.com` and password `password2`:

```
$ telnet 127.0.0.1 1143
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.
* OK [CAPABILITY IDLE IMAP4rev1 MOVE UIDPLUS UNSELECT] gluon session ID 1
tag login user1@example.com password1
tag OK [CAPABILITY IDLE IMAP4rev1 MOVE UIDPLUS UNSELECT] (^_^)

...

tag login alias1@example.com password1
tag OK [CAPABILITY IDLE IMAP4rev1 MOVE UIDPLUS UNSELECT] (^_^)

...

tag login user2@example.com password2
tag OK [CAPABILITY IDLE IMAP4rev1 MOVE UIDPLUS UNSELECT] (^_^)

...

tag login alias2@example.com password2
tag OK [CAPABILITY IDLE IMAP4rev1 MOVE UIDPLUS UNSELECT] (^_^)
```

The demo accounts contain no messages. You can connect an IMAP client (e.g. thunderbird) and use it to copy in
messages from another mail server.


# Changing DB schema

Do not forget to re-generate ent code after changing the DB schema in `./internal/db/ent/schema`.

```
pushd ./internal/db/ent && go generate . && popd

