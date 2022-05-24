package imap

import (
	"fmt"
	"net/mail"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type parList []fmt.Stringer

func (l parList) String() string {
	if len(l) == 0 {
		return "NIL"
	}

	var res []string

	for _, item := range l {
		res = append(res, item.String())
	}

	return fmt.Sprintf("(%v)", strings.Join(res, " "))
}

func (l *parList) add(v any) *parList {
	switch v := v.(type) {
	case string:
		*l = append(*l, nilString(v))

	case int:
		*l = append(*l, numString(v))

	case fmt.Stringer:
		*l = append(*l, v)

	case []fmt.Stringer:
		*l = append(*l, concatList(v))

	case map[string]string:
		keys := maps.Keys(v)

		slices.Sort(keys)

		var params parList

		for _, key := range keys {
			params = append(params, nilString(key), nilString(v[key]))
		}

		*l = append(*l, params)

	case []*mail.Address:
		var addrList parList

		for _, addr := range v {
			var user, domain string

			if split := strings.Split(addr.Address, "@"); len(split) == 2 {
				user, domain = split[0], split[1]
			}

			var fields parList

			fields.
				add(addr.Name).
				add("").
				add(user).
				add(domain)

			addrList.add(fields)
		}

		*l = append(*l, addrList)

	default:
		panic(v)
	}

	return l
}

type nilString string

func (s nilString) String() string {
	if s == "" {
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
	var res string

	for _, item := range l {
		res += item.String()
	}

	return res
}
