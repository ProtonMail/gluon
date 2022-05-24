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

func (s *itemMessages) String() string {
	return fmt.Sprintf("MESSAGES %v", s.n)
}
