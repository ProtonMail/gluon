package response

import "fmt"

type itemMessages struct {
	n int
}

func ItemMessages(n int) *itemMessages {
	return &itemMessages{
		n: n,
	}
}

func (s *itemMessages) Strings() (raw string, _ string) {
	raw = fmt.Sprintf("MESSAGES %v", s.n)
	return raw, raw
}
