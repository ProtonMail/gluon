package benchmarks

import (
	"fmt"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
	"github.com/emersion/go-imap"
)

// ParallelSeqSet contains a list of sequence sets which can be used by the benchmarks. Use one of the several new
// functions to initialize the state. Internally it holds one list per expected worker.
type ParallelSeqSet struct {
	seqSets [][]*imap.SeqSet
}

func (p *ParallelSeqSet) Get(i uint) []*imap.SeqSet {
	return p.seqSets[i]
}

// NewParallelSeqSetFromFile load the sequence sets from a file. The same sequence set will be assigned to
// all workers.
func NewParallelSeqSetFromFile(path string, numWorkers uint) (*ParallelSeqSet, error) {
	list, err := utils.SequenceListFromFile(path)
	if err != nil {
		return nil, err
	}

	seqSets := make([][]*imap.SeqSet, numWorkers)
	for i := uint(0); i < numWorkers; i++ {
		seqSets[i] = list
	}

	return &ParallelSeqSet{
		seqSets: seqSets,
	}, nil
}

// NewParallelSeqSetRandom generates count random sequence set for each worker. If generateIntervals is set to true,
// it will generate intervals rather than a single number.
func NewParallelSeqSetRandom(count uint32, numWorkers uint, generateIntervals bool) *ParallelSeqSet {
	lists := make([][]*imap.SeqSet, numWorkers)

	for i := uint(0); i < numWorkers; i++ {
		list := make([]*imap.SeqSet, 0, count)

		for r := uint32(0); r < count; r++ {
			var seqSet *imap.SeqSet
			if !generateIntervals {
				seqSet = utils.RandomSequenceSetNum(count)
			} else {
				seqSet = utils.RandomSequenceSetRange(count)
			}

			list = append(list, seqSet)
		}

		lists[i] = list
	}

	return &ParallelSeqSet{seqSets: lists}
}

// NewParallelSeqSetAll generates once sequence set for each worker which covers everything (1:*).
func NewParallelSeqSetAll(numWorkers uint) *ParallelSeqSet {
	lists := make([][]*imap.SeqSet, numWorkers)
	for i := uint(0); i < numWorkers; i++ {
		lists[i] = []*imap.SeqSet{utils.NewSequenceSetAll()}
	}

	return &ParallelSeqSet{seqSets: lists}
}

// NewParallelSeqSet generates a parallel SeqSet based on the following conditions:
// * If a listFile is not empty, it will load the sequence sets from that file.
// * If generateAll is set to true, it will call NewParallelSeqSetAll.
// * If none of the above are valid it will generate random collection of sequence sets which can be single or intervals
//   based on whether generateIntervals is set to true.
func NewParallelSeqSet(count uint32, numWorkers uint, listFile string, generateAll bool, generateIntervals bool) (*ParallelSeqSet, error) {
	if count == 0 {
		return nil, fmt.Errorf("count can not be 0")
	}

	if len(listFile) != 0 {
		return NewParallelSeqSetFromFile(listFile, numWorkers)
	} else if generateAll {
		return NewParallelSeqSetAll(numWorkers), nil
	} else {
		return NewParallelSeqSetRandom(count, numWorkers, generateIntervals), nil
	}
}
