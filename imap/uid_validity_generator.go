package imap

import (
	"fmt"
	"sync/atomic"
	"time"
)

type UIDValidityGenerator interface {
	Generate() (UID, error)
}

type EpochUIDValidityGenerator struct {
	epochStart time.Time
	lastUID    uint32
}

func NewEpochUIDValidityGenerator(epochStart time.Time) *EpochUIDValidityGenerator {
	return &EpochUIDValidityGenerator{
		epochStart: epochStart,
	}
}

func DefaultEpochUIDValidityGenerator() *EpochUIDValidityGenerator {
	return NewEpochUIDValidityGenerator(time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC))
}

func (e *EpochUIDValidityGenerator) Generate() (UID, error) {
	timeStamp := uint64(time.Now().Sub(e.epochStart).Seconds())
	if timeStamp > uint64(0xFFFFFFFF) {
		return 0, fmt.Errorf("failed to generate uid validity, interval exceeded maximum capacity")
	}

	timeStampU32 := uint32(timeStamp)

	// This loops is here to ensure that two successive calls to Generate that happen during the same second
	// can still generate unique values. To avoid waiting another second until the values are different,
	// we keep bumping the last generated value until it is greater than the last generated value.
	for {
		lastGenerated := atomic.LoadUint32(&e.lastUID)

		// Not enough time elapsed between the last time
		if lastGenerated >= timeStampU32 {
			if timeStampU32 == 0xFFFFFFFF {
				return 0, fmt.Errorf("failed to generate uid validity, interval exceeded maximum capacity")
			}

			timeStampU32 += 1

			continue
		}

		if !atomic.CompareAndSwapUint32(&e.lastUID, lastGenerated, timeStampU32) {
			continue
		}

		return UID(timeStampU32), nil
	}
}

type IncrementalUIDValidityGenerator struct {
	counter uint32
}

func (i *IncrementalUIDValidityGenerator) Generate() (UID, error) {
	return UID(atomic.AddUint32(&i.counter, 1)), nil
}

func (i *IncrementalUIDValidityGenerator) GetValue() UID {
	return UID(atomic.LoadUint32(&i.counter))
}

func NewIncrementalUIDValidityGenerator() *IncrementalUIDValidityGenerator {
	return &IncrementalUIDValidityGenerator{}
}

type FixedUIDValidityGenerator struct {
	Value UID
}

func (f FixedUIDValidityGenerator) Generate() (UID, error) {
	return f.Value, nil
}

func NewFixedUIDValidityGenerator(value UID) *FixedUIDValidityGenerator {
	return &FixedUIDValidityGenerator{Value: value}
}
