package imap_benchmarks

import (
	"fmt"
	"math/rand"

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
	list, err := SequenceListFromFile(path)
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

// NewParallelSeqSetExpunge generates sequence ids or intervals that can be used in cases where the data is expunged
// and/or moved from the original inbox. It also makes sure that concurrent workers can't overlap to avoid operations
// on messages that no longer exist.
func NewParallelSeqSetExpunge(count uint32, numWorkers uint, generateIntervals, uid bool) *ParallelSeqSet {
	lists := make([][]*imap.SeqSet, numWorkers)
	workerSplit := count / uint32(numWorkers)
	available := make([]uint32, count)

	for r := uint32(0); r < count; r++ {
		available[r] = r + 1
	}

	for i := uint(0); i < numWorkers; i++ {
		available := available[(uint32(i) * workerSplit):(uint32(i+1) * workerSplit)]
		list := make([]*imap.SeqSet, 0, workerSplit)

		if generateIntervals {
			const maxIntervalRange = uint32(40)
			for len(available) > 0 {
				intervalRange := rand.Uint32() % maxIntervalRange
				itemsLeft := uint32(len(available))
				index := rand.Uint32() % itemsLeft

				if index > intervalRange {
					index -= intervalRange
				} else {
					index = 0
				}

				if index+intervalRange >= itemsLeft {
					intervalRange = itemsLeft - index
				}

				seqSet := &imap.SeqSet{}
				if uid {
					seqSet.AddRange(available[index], available[index+intervalRange-1])
				} else {
					seqSet.AddRange(index+1, index+intervalRange)
				}

				list = append(list, seqSet)
				available = append(available[:index], available[index+intervalRange:]...)
			}
		} else {
			count := uint32(len(available))
			for r := uint32(0); r < count; r++ {
				index := rand.Uint32() % (count - r)
				seqSet := &imap.SeqSet{}
				if uid {
					tmp := available[index]
					available[index] = available[count-r-1]
					seqSet.AddNum(tmp)
				} else {
					seqSet.AddNum(index)
				}
				list = append(list, seqSet)
			}
		}

		lists[i] = list
	}

	return &ParallelSeqSet{seqSets: lists}
}

// NewParallelSeqSetRandom generates count random sequence set for each worker. If generateIntervals is set to true,
// it will generate intervals rather than a single number. If randomDrain is set to true it will generate unique
// values that eventually exhaust the problem space.
func NewParallelSeqSetRandom(count uint32, numWorkers uint, generateIntervals, randomDrain, uid bool) *ParallelSeqSet {
	lists := make([][]*imap.SeqSet, numWorkers)

	for i := uint(0); i < numWorkers; i++ {
		list := make([]*imap.SeqSet, 0, count)

		if randomDrain {
			available := make([]uint32, count)
			for r := uint32(0); r < count; r++ {
				available[r] = r + 1
			}

			if generateIntervals {
				const maxIntervalRange = uint32(40)
				for len(available) > 0 {
					intervalRange := rand.Uint32() % maxIntervalRange
					itemsLeft := uint32(len(available))
					index := rand.Uint32() % itemsLeft

					if index > intervalRange {
						index -= intervalRange
					} else {
						index = 0
					}

					if index+intervalRange >= itemsLeft {
						intervalRange = itemsLeft - index
					}

					seqSet := &imap.SeqSet{}
					if uid {
						seqSet.AddRange(available[index], available[index+intervalRange-1])
					} else {
						seqSet.AddRange(index+1, index+intervalRange)
					}

					list = append(list, seqSet)
					available = append(available[:index], available[index+intervalRange:]...)
				}
			} else {
				for r := uint32(0); r < count; r++ {
					index := rand.Uint32() % (count - r)
					seqSet := &imap.SeqSet{}
					if uid {
						tmp := available[index]
						available[index] = available[count-r-1]
						seqSet.AddNum(tmp)
					} else {
						seqSet.AddNum(index)
					}
					list = append(list, seqSet)
				}
			}
		} else {
			for r := uint32(0); r < count; r++ {
				var seqSet *imap.SeqSet
				if !generateIntervals {
					seqSet = RandomSequenceSetNum(count)
				} else {
					seqSet = RandomSequenceSetRange(count)
				}

				list = append(list, seqSet)
			}
		}

		lists[i] = list
	}

	return &ParallelSeqSet{seqSets: lists}
}

// NewParallelSeqSetAll generates once sequence set for each worker which covers everything (1:*).
func NewParallelSeqSetAll(numWorkers uint) *ParallelSeqSet {
	lists := make([][]*imap.SeqSet, numWorkers)
	for i := uint(0); i < numWorkers; i++ {
		lists[i] = []*imap.SeqSet{NewSequenceSetAll()}
	}

	return &ParallelSeqSet{seqSets: lists}
}

// NewParallelSeqSet generates a parallel SeqSet based on the following conditions:
//   - If a listFile is not empty, it will load the sequence sets from that file.
//   - If generateAll is set to true, it will call NewParallelSeqSetAll.
//   - If none of the above are valid it will generate random collection of sequence sets which can be single or intervals
//     based on whether generateIntervals is set to true.
//     If randomDrain is set to true, it will generate non repeating sequences. E.g. Useful for move or delete benchmarks.
//     If uid is set to true, it will assume the values are UIDs rather than sequence IDs.
func NewParallelSeqSet(count uint32, numWorkers uint, listFile string, generateAll, generateIntervals, randomDrain, uid bool) (*ParallelSeqSet, error) {
	if count == 0 {
		return nil, fmt.Errorf("count can not be 0")
	}

	if len(listFile) != 0 {
		return NewParallelSeqSetFromFile(listFile, numWorkers)
	} else if generateAll {
		return NewParallelSeqSetAll(numWorkers), nil
	} else {
		return NewParallelSeqSetRandom(count, numWorkers, generateIntervals, randomDrain, uid), nil
	}
}
