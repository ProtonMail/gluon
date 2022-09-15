package imap

import (
	"bytes"
	"errors"
	"fmt"
	"mime"
	"strings"

	"github.com/ProtonMail/gluon/rfc822"
)

func Structure(section *rfc822.Section, ext bool) (string, error) {
	res, err := structure(section, ext)
	if err != nil {
		return "", err
	}

	return res.String(), nil
}

func structure(section *rfc822.Section, ext bool) (fmt.Stringer, error) {
	if len(section.Children()) == 0 {
		return singlePartStructure(section, ext)
	}

	var fields parList

	children, err := childStructures(section, ext)
	if err != nil {
		return nil, err
	}

	fields.addStringers(children)

	header, err := section.ParseHeader()
	if err != nil {
		return nil, err
	}

	_, mimeSubType, mimeParams, err := getMIMEInfo(header)
	if err != nil {
		return nil, err
	}

	fields.addString(mimeSubType)

	if ext {
		fields.
			addMap(mimeParams).
			addStringer(getDispInfo(header)).
			addString(header.Get("Content-Language")).
			addString(header.Get("Content-Location"))
	}

	return fields, nil
}

func singlePartStructure(section *rfc822.Section, ext bool) (fmt.Stringer, error) {
	header, err := section.ParseHeader()
	if err != nil {
		return nil, err
	}

	var fields parList

	mimeType, mimeSubType, mimeParams, err := getMIMEInfo(header)
	if err != nil {
		return nil, err
	}

	fields.
		addString(mimeType).
		addString(mimeSubType).
		addMap(mimeParams).
		addString(header.Get("Content-Id")).
		addString(header.Get("Content-Description")).
		addString(header.Get("Content-Transfer-Encoding")).
		addNumber(len(section.Body()))

	if mimeType == "message" && mimeSubType == "rfc822" {
		child := rfc822.Parse(section.Body())

		header, err := child.ParseHeader()
		if err != nil {
			return nil, err
		}

		envelope, err := envelope(header)
		if err != nil {
			return nil, err
		}

		body, err := structure(child, ext)
		if err != nil {
			return nil, err
		}

		fields.addStringer(envelope).addStringer(body)
	}

	if mimeType == "text" || (mimeType == "message" && mimeSubType == "rfc822") {
		fields.addNumber(countLines(section.Body()))
	}

	if ext {
		fields.
			addString(header.Get("Content-MD5")).
			addStringer(getDispInfo(header)).
			addString(header.Get("Content-Language")).
			addString(header.Get("Content-Location"))
	}

	return fields, nil
}

func childStructures(section *rfc822.Section, ext bool) ([]fmt.Stringer, error) {
	var children []fmt.Stringer

	for _, child := range section.Children() {
		structure, err := structure(child, ext)
		if err != nil {
			return nil, err
		}

		children = append(children, structure)
	}

	return children, nil
}

func getMIMEInfo(header *rfc822.Header) (string, string, map[string]string, error) {
	contentType, contentTypeParams, err := rfc822.ParseContentType(header.Get("Content-Type"))
	if err != nil {
		return "", "", nil, err
	}

	split := strings.Split(contentType, "/")
	if len(split) != 2 {
		return "", "", nil, errors.New("malformed MIME type")
	}

	return split[0], split[1], contentTypeParams, nil
}

func getDispInfo(header *rfc822.Header) fmt.Stringer {
	var fields parList

	if contentDisp, contentDispParams, err := mime.ParseMediaType(header.Get("Content-Disposition")); err == nil {
		fields.addString(contentDisp).addMap(contentDispParams)
	}

	return fields
}

func countLines(b []byte) int {
	lines := 0
	remaining := b
	separator := []byte{'\n'}

	for len(remaining) != 0 {
		index := bytes.Index(remaining, separator)
		if index < 0 {
			lines++
			break
		}

		lines++

		remaining = remaining[index+1:]
	}

	return lines
}
