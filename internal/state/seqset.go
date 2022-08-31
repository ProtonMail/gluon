package state

import (
	"fmt"

	"github.com/ProtonMail/gluon/internal/parser/proto"
)

func toSeqSet(set *proto.SequenceSet) ([][]string, error) {
	var seqRanges [][]string

	for _, item := range set.GetItems() {
		var seqRange []string

		switch item := item.GetItem().(type) {
		case *proto.SequenceItem_Number:
			seqRange = append(seqRange, item.Number)

		case *proto.SequenceItem_Range:
			seqRange = append(seqRange, item.Range.GetBegin(), item.Range.GetEnd())

		default:
			return nil, fmt.Errorf("bad sequence range")
		}

		seqRanges = append(seqRanges, seqRange)
	}

	return seqRanges, nil
}
