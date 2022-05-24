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

func (s *itemRecent) String() string {
	return fmt.Sprintf("RECENT %v", s.n)
}
