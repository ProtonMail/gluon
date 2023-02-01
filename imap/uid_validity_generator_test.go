package imap

import (
	"testing"
	"time"

	"github.com/bradenaw/juniper/parallel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestEpochUIDValidityGenerator_Generate(t *testing.T) {
	generator := DefaultEpochUIDValidityGenerator()

	const UIDCount = 10

	var uids = make([]UID, UIDCount)

	for i := 0; i < UIDCount; i++ {
		uid, err := generator.Generate()
		require.NoError(t, err)

		uids[i] = uid
	}

	time.Sleep(10 * time.Second)

	uid, err := generator.Generate()
	require.NoError(t, err)

	for i := 0; i < UIDCount-1; i++ {
		assert.Less(t, uids[i], uids[i+1])
	}

	assert.Greater(t, uid, uids[UIDCount-1])
}

func TestEpochUIDValidityGenerator_GenerateParallel(t *testing.T) {
	generator := DefaultEpochUIDValidityGenerator()

	const UIDCount = 1000

	var uids = make([]UID, UIDCount)

	parallel.Do(0, UIDCount, func(i int) {
		uid, err := generator.Generate()
		require.NoError(t, err)
		uids[i] = uid
	})

	slices.Sort(uids)

	for i := 0; i < UIDCount-1; i++ {
		assert.Less(t, uids[i], uids[i+1])
	}
}
