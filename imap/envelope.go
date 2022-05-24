package imap

import (
	"fmt"
	"net/mail"

	"github.com/ProtonMail/gluon/rfc822"
)

func Envelope(header *rfc822.Header) (string, error) {
	res, err := envelope(header)
	if err != nil {
		return "", err
	}

	return res.String(), nil
}

func envelope(header *rfc822.Header) (fmt.Stringer, error) {
	var fields parList

	fields.
		add(header.Get("Date")).
		add(header.Get("Subject"))

	if !header.Has("From") {
		fields.add("")
	} else {
		fields.add(tryParseAddressList(header.Get("From")))
	}

	switch {
	case header.Has("Sender"):
		fields.add(tryParseAddressList(header.Get("Sender")))

	case header.Has("From"):
		fields.add(tryParseAddressList(header.Get("From")))

	default:
		fields.add("")
	}

	switch {
	case header.Has("Reply-To"):
		fields.add(tryParseAddressList(header.Get("Reply-To")))

	case header.Has("From"):
		fields.add(tryParseAddressList(header.Get("From")))

	default:
		fields.add("")
	}

	if !header.Has("To") {
		fields.add("")
	} else {
		fields.add(tryParseAddressList(header.Get("To")))
	}

	if !header.Has("Cc") {
		fields.add("")
	} else {
		fields.add(tryParseAddressList(header.Get("Cc")))
	}

	if !header.Has("Bcc") {
		fields.add("")
	} else {
		fields.add(tryParseAddressList(header.Get("Bcc")))
	}

	if !header.Has("In-Reply-To") {
		fields.add("")
	} else {
		fields.add(tryParseAddressList(header.Get("In-Reply-To")))
	}

	fields.add(header.Get("Message-Id"))

	return fields, nil
}

// TODO: Should use RFC5322 package here but it's too slow... sad.
func tryParseAddressList(val string) []*mail.Address {
	if addr, err := mail.ParseAddressList(val); err == nil {
		return addr
	}

	return []*mail.Address{{Address: val}}
}
