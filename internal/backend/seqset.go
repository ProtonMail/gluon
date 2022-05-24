package backend

import "github.com/ProtonMail/gluon/internal/parser/proto"

func toSeqSet(set *proto.SequenceSet) [][]string {
	var seqRanges [][]string

	for _, item := range set.GetItems() {
		var seqRange []string

		switch item := item.GetItem().(type) {
		case *proto.SequenceItem_Number:
			seqRange = append(seqRange, item.Number)

		case *proto.SequenceItem_Range:
			seqRange = append(seqRange, item.Range.GetBegin(), item.Range.GetEnd())

		default:
			panic("bad sequence range")
		}

		seqRanges = append(seqRanges, seqRange)
	}

	return seqRanges
}
