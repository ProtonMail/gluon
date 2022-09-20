package imap

import (
	"net/mail"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type parListWriter interface {
	writeString(string)
	writeByte(byte)
}

type singleParListWriter struct {
	b *strings.Builder
}

func (s *singleParListWriter) writeString(v string) {
	s.b.WriteString(v)
}

func (s *singleParListWriter) writeByte(v byte) {
	s.b.WriteByte(v)
}

type dualParListWriter struct {
	b1 *strings.Builder
	b2 *strings.Builder
}

func (d *dualParListWriter) writeString(v string) {
	d.b1.WriteString(v)
	d.b2.WriteString(v)
}

func (d *dualParListWriter) writeByte(v byte) {
	d.b1.WriteByte(v)
	d.b2.WriteByte(v)
}

func (d *dualParListWriter) toSingleWriterFrom1st() parListWriter {
	return &singleParListWriter{b: d.b1}
}

func (d *dualParListWriter) toSingleWriterFrom2nd() parListWriter {
	return &singleParListWriter{b: d.b2}
}

type paramList struct {
	firstItem bool
}

func newParamListWithGroup(writer parListWriter) paramList {
	writer.writeByte('(')

	return paramList{
		firstItem: true,
	}
}

func newParamListWithoutGroup() paramList {
	return paramList{
		firstItem: true,
	}
}

func (c *paramList) newChildList(writer parListWriter) paramList {
	c.firstItem = false
	return newParamListWithGroup(writer)
}

func (c *paramList) finish(writer parListWriter) {
	writer.writeByte(')')
}

func (c *paramList) onWrite(writer parListWriter) {
	if !c.firstItem {
		writer.writeByte(' ')
	}

	c.firstItem = false
}

func (c *paramList) addString(writer parListWriter, v string) *paramList {
	c.onWrite(writer)

	var str string

	if len(v) == 0 {
		str = "NIL"
	} else {
		str = strconv.Quote(v)
	}

	writer.writeString(str)

	return c
}

func (c *paramList) addNumber(writer parListWriter, v int) *paramList {
	c.onWrite(writer)

	str := strconv.Itoa(v)

	writer.writeString(str)

	return c
}

func (c *paramList) addMap(writer parListWriter, v map[string]string) *paramList {
	c.onWrite(writer)

	keys := maps.Keys(v)

	slices.Sort(keys)

	child := c.newChildList(writer)

	for _, key := range keys {
		child.addString(writer, key).addString(writer, v[key])
	}

	child.finish(writer)

	return c
}

func (c *paramList) addAddresses(writer parListWriter, v []*mail.Address) *paramList {
	c.onWrite(writer)

	child := c.newChildList(writer)

	for _, addr := range v {
		var user, domain string

		if split := strings.Split(addr.Address, "@"); len(split) == 2 {
			user, domain = split[0], split[1]
		}

		fields := child.newChildList(writer)

		fields.
			addString(writer, addr.Name).
			addString(writer, "").
			addString(writer, user).
			addString(writer, domain)

		fields.finish(writer)
	}

	child.finish(writer)

	return c
}
