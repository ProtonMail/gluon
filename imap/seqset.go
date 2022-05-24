package imap

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

type SeqVal struct {
	Begin, End int
}

func (seqval SeqVal) canCombine(val int) bool {
	return val == seqval.End+1
}

func (seqval SeqVal) String() string {
	if seqval.End > seqval.Begin {
		return fmt.Sprintf("%v:%v", seqval.Begin, seqval.End)
	}

	return strconv.Itoa(seqval.End)
}

type SeqSet []SeqVal

func NewSeqSet(set []int) SeqSet {
	slices.Sort(set)

	var res SeqSet

	for _, val := range set {
		if n := len(res); n > 0 {
			if res[n-1].canCombine(val) {
				res[n-1].End = val
			} else {
				res = append(res, SeqVal{Begin: val, End: val})
			}
		} else {
			res = append(res, SeqVal{Begin: val, End: val})
		}
	}

	return res
}

func (set SeqSet) String() string {
	var res []string

	for _, val := range set {
		res = append(res, val.String())
	}

	return strings.Join(res, ",")
}
