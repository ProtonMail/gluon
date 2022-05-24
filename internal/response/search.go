package response

import (
	"strconv"

	"golang.org/x/exp/slices"
)

type search struct {
	seqs []int
}

func Search(seqs ...int) *search {
	slices.Sort(seqs)

	return &search{
		seqs: seqs,
	}
}

func (r *search) Send(s Session) error {
	return s.WriteResponse(r.String())
}

func (r *search) String() string {
	parts := []string{"*", "SEARCH"}

	if len(r.seqs) > 0 {
		var seqs []string

		for _, seq := range r.seqs {
			seqs = append(seqs, strconv.Itoa(seq))
		}

		parts = append(parts, join(seqs))
	}

	return join(parts)
}
