package imap

import (
	"fmt"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"net/mail"
	"strconv"
	"strings"
)

type parList []fmt.Stringer

func (l parList) String() string {
	itemsLen := len(l)
	if itemsLen == 0 {
		return "NIL"
	}

	builder := strings.Builder{}
	builder.WriteRune('(')

	builder.WriteString(l[0].String())

	for _, item := range l[1:] {
		builder.WriteRune(' ')
		builder.WriteString(item.String())
	}

	builder.WriteRune(')')

	return builder.String()
}

func (l *parList) addString(v string) *parList {
	*l = append(*l, nilString(v))

	return l
}

func (l *parList) addNumber(v int) *parList {
	*l = append(*l, numString(v))

	return l
}

func (l *parList) addStringer(v fmt.Stringer) *parList {
	*l = append(*l, v)

	return l
}

func (l *parList) addStringers(v []fmt.Stringer) *parList {
	*l = append(*l, concatList(v))

	return l
}

func (l *parList) addMap(v map[string]string) *parList {
	keys := maps.Keys(v)

	slices.Sort(keys)

	params := make(parList, 0, len(keys))

	for _, key := range keys {
		params = append(params, nilString(key), nilString(v[key]))
	}

	*l = append(*l, params)

	return l
}

func (l *parList) addAddresses(v []*mail.Address) *parList {
	var addrList parList

	for _, addr := range v {
		var user, domain string

		if split := strings.Split(addr.Address, "@"); len(split) == 2 {
			user, domain = split[0], split[1]
		}

		fields := make(parList, 0, 4)

		fields.
			addString(addr.Name).
			addString("").
			addString(user).
			addString(domain)

		addrList.addStringer(fields)
	}

	*l = append(*l, addrList)

	return l
}

type nilString string

func (s nilString) String() string {
	if len(s) == 0 {
		return "NIL"
	}

	return strconv.Quote(string(s))
}

type numString int

func (n numString) String() string {
	return strconv.Itoa(int(n))
}

type concatList []fmt.Stringer

func (l concatList) String() string {
	builder := strings.Builder{}
	for _, item := range l {
		builder.WriteString(item.String())
	}

	return builder.String()
}
