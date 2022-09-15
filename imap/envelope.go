package imap

import (
	"net/mail"
	"strings"

	"github.com/ProtonMail/gluon/rfc822"
)

func Envelope(header *rfc822.Header) (string, error) {
	builder := strings.Builder{}
	writer := singleParListWriter{b: &builder}
	paramList := newParamListWithoutGroup()

	if err := envelope(header, &paramList, &writer); err != nil {
		return "", err
	}

	return builder.String(), nil
}

func envelope(header *rfc822.Header, c *paramList, writer parListWriter) error {
	fields := c.newChildList(writer)
	defer fields.finish(writer)

	fields.
		addString(writer, header.Get("Date")).
		addString(writer, header.Get("Subject"))

	if v, ok := header.GetChecked("From"); !ok {
		fields.addString(writer, "")
	} else {
		fields.addAddresses(writer, tryParseAddressList(v))
	}

	if v, ok := header.GetChecked("Sender"); ok {
		fields.addAddresses(writer, tryParseAddressList(v))
	} else if v, ok := header.GetChecked("From"); ok {
		fields.addAddresses(writer, tryParseAddressList(v))
	} else {
		fields.addString(writer, "")
	}

	if v, ok := header.GetChecked("Reply-To"); ok {
		fields.addAddresses(writer, tryParseAddressList(v))
	} else if v, ok := header.GetChecked("From"); ok {
		fields.addAddresses(writer, tryParseAddressList(v))
	} else {
		fields.addString(writer, "")
	}

	if v, ok := header.GetChecked("To"); !ok {
		fields.addString(writer, "")
	} else {
		fields.addAddresses(writer, tryParseAddressList(v))
	}

	if v, ok := header.GetChecked("Cc"); !ok {
		fields.addString(writer, "")
	} else {
		fields.addAddresses(writer, tryParseAddressList(v))
	}

	if v, ok := header.GetChecked("Bcc"); !ok {
		fields.addString(writer, "")
	} else {
		fields.addAddresses(writer, tryParseAddressList(v))
	}

	fields.addString(writer, header.Get("In-Reply-To"))

	fields.addString(writer, header.Get("Message-Id"))

	return nil
}

// TODO: Should use RFC5322 package here but it's too slow... sad.
func tryParseAddressList(val string) []*mail.Address {
	if addr, err := mail.ParseAddressList(val); err == nil {
		return addr
	}

	return []*mail.Address{{Address: val}}
}
