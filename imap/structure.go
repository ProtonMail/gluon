package imap

import (
	"bytes"
	"strings"

	"github.com/ProtonMail/gluon/rfc822"
)

func Structure(section *rfc822.Section) (string, string, error) {
	bodyBuilder := strings.Builder{}
	structureBuilder := strings.Builder{}

	writer := dualParListWriter{b1: &bodyBuilder, b2: &structureBuilder}

	c := newParamListWithGroup(&writer)
	if err := structure(section, &c, &writer); err != nil {
		return "", "", err
	}

	c.finish(&writer)

	body := bodyBuilder.String()
	structure := structureBuilder.String()

	return body, structure, nil
}

func structure(section *rfc822.Section, fields *paramList, writer *dualParListWriter) error {
	children, err := section.Children()
	if err != nil {
		return err
	}

	if len(children) == 0 {
		return singlePartStructure(section, fields, writer)
	}

	if err := childStructures(section, fields, writer); err != nil {
		return err
	}

	header, err := section.ParseHeader()
	if err != nil {
		return err
	}

	_, mimeSubType, mimeParams, err := getMIMEInfo(section)
	if err != nil {
		return err
	}

	fields.addString(writer, mimeSubType)

	extWriter := writer.toSingleWriterFrom2nd()
	fields.addMap(extWriter, mimeParams)
	addDispInfo(fields, extWriter, header)
	fields.addString(extWriter, header.Get("Content-Language")).
		addString(extWriter, header.Get("Content-Location"))

	return nil
}

func singlePartStructure(section *rfc822.Section, fields *paramList, writer *dualParListWriter) error {
	header, err := section.ParseHeader()
	if err != nil {
		return err
	}

	mimeType, mimeSubType, mimeParams, err := getMIMEInfo(section)
	if err != nil {
		return err
	}

	fields.
		addString(writer, mimeType).
		addString(writer, mimeSubType).
		addMap(writer, mimeParams).
		addString(writer, header.Get("Content-Id")).
		addString(writer, header.Get("Content-Description")).
		addString(writer, header.Get("Content-Transfer-Encoding")).
		addNumber(writer, len(section.Body()))

	if mimeType == "message" && mimeSubType == "rfc822" {
		child := rfc822.Parse(section.Body())

		header, err := child.ParseHeader()
		if err != nil {
			return err
		}

		writer.writeByte(' ')

		if err := envelope(header, fields, writer); err != nil {
			return err
		}

		cstruct := fields.newChildList(writer)

		if err := structure(child, &cstruct, writer); err != nil {
			return err
		}

		cstruct.finish(writer)
	}

	if mimeType == "text" || (mimeType == "message" && mimeSubType == "rfc822") {
		fields.addNumber(writer, countLines(section.Body()))
	}

	extWriter := writer.toSingleWriterFrom2nd()
	fields.addString(extWriter, header.Get("Content-MD5"))
	addDispInfo(fields, extWriter, header)
	fields.addString(extWriter, header.Get("Content-Language")).
		addString(extWriter, header.Get("Content-Location"))

	return nil
}

func childStructures(section *rfc822.Section, c *paramList, writer *dualParListWriter) error {
	children, err := section.Children()
	if err != nil {
		return err
	}

	for _, child := range children {
		cl := c.newChildList(writer)

		if err := structure(child, &cl, writer); err != nil {
			return err
		}

		cl.finish(writer)
	}

	return nil
}

func getMIMEInfo(section *rfc822.Section) (string, string, map[string]string, error) {
	mimeType, mimeParams, err := section.ContentType()
	if err != nil {
		return "", "", nil, err
	}

	return mimeType.Type(), mimeType.SubType(), mimeParams, nil
}

func addDispInfo(c *paramList, writer parListWriter, header *rfc822.Header) {
	if contentDisp, contentDispParams, err := rfc822.ParseMediaType(header.Get("Content-Disposition")); err == nil {
		writer.writeByte(' ')
		fields := c.newChildList(writer)
		fields.addString(writer, contentDisp).addMap(writer, contentDispParams)
		fields.finish(writer)
	} else {
		c.addString(writer, "")
	}
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
