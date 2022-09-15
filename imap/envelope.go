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
		addString(header.Get("Date")).
		addString(header.Get("Subject"))

	if v, ok := header.GetChecked("From"); !ok {
		fields.addString("")
	} else {
		fields.addAddresses(tryParseAddressList(v))
	}

	if v, ok := header.GetChecked("Sender"); ok {
		fields.addAddresses(tryParseAddressList(v))
	} else if v, ok := header.GetChecked("From"); ok {
		fields.addAddresses(tryParseAddressList(v))
	} else {
		fields.addString("")
	}

	if v, ok := header.GetChecked("Reply-To"); ok {
		fields.addAddresses(tryParseAddressList(v))
	} else if v, ok := header.GetChecked("From"); ok {
		fields.addAddresses(tryParseAddressList(v))
	} else {
		fields.addString("")
	}

	if v, ok := header.GetChecked("To"); !ok {
		fields.addString("")
	} else {
		fields.addAddresses(tryParseAddressList(v))
	}

	if v, ok := header.GetChecked("Cc"); !ok {
		fields.addString("")
	} else {
		fields.addAddresses(tryParseAddressList(v))
	}

	if v, ok := header.GetChecked("Bcc"); !ok {
		fields.addString("")
	} else {
		fields.addAddresses(tryParseAddressList(v))
	}

	fields.addString(header.Get("In-Reply-To"))

	fields.addString(header.Get("Message-Id"))

	return fields, nil
}

// TODO: Should use RFC5322 package here but it's too slow... sad.
func tryParseAddressList(val string) []*mail.Address {
	if addr, err := mail.ParseAddressList(val); err == nil {
		return addr
	}

	return []*mail.Address{{Address: val}}
}
