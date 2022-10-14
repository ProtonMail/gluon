package response

import "fmt"

type itemRecent struct {
	n int
}

func ItemRecent(n int) *itemRecent {
	return &itemRecent{
		n: n,
	}
}

func (s *itemRecent) Strings() (raw string, _ string) {
	raw = fmt.Sprintf("RECENT %v", s.n)
	return raw, raw
}
