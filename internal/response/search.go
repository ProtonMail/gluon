package response

import (
	"strconv"

	"golang.org/x/exp/slices"
)

type search struct {
	seqs []uint32
}

func Search(seqs ...uint32) *search {
	slices.Sort(seqs)

	return &search{
		seqs: seqs,
	}
}

func (r *search) Send(s Session) error {
	return s.WriteResponse(r)
}

func (r *search) Strings() (raw string, _ string) {
	parts := []string{"*", "SEARCH"}

	if len(r.seqs) > 0 {
		var seqs []string

		for _, seq := range r.seqs {
			seqs = append(seqs, strconv.Itoa(int(seq)))
		}

		parts = append(parts, join(seqs))
	}

	raw = join(parts)

	return raw, raw
}
