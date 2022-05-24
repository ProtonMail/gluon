package imap

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFlagSet(t *testing.T) {
	fs := NewFlagSet()
	require.Equal(t, 0, fs.Len())

	fs = NewFlagSet("flag1")
	require.Equal(t, 1, fs.Len())
	require.ElementsMatch(t, fs.ToSlice(), []string{"flag1"})

	fs = NewFlagSet("flag1", "flag1", "FLAG1")
	require.Equal(t, 1, fs.Len())
	require.ElementsMatch(t, fs.ToSlice(), []string{"flag1"})

	fs = NewFlagSet("flag1", "FLAG2", "flag2", "FLAG1")
	require.Equal(t, 2, fs.Len())
	require.ElementsMatch(t, fs.ToSlice(), []string{"flag1", "FLAG2"})
}

func TestFlagSet_Contains(t *testing.T) {
	fs := NewFlagSet("flag1", "flag2", "flag3", "flag4")
	require.Equal(t, 4, fs.Len())

	require.False(t, NewFlagSet().Contains("flag1"))
	require.True(t, fs.Contains("flag1"))
	require.True(t, fs.Contains("FLAG1"))
	require.True(t, fs.Contains("flAg2"))
	require.True(t, fs.Contains("flag3"))
	require.True(t, fs.Contains("flag4"))
	require.False(t, fs.Contains("flag5"))
	require.False(t, fs.Contains("flag4 "))
	require.False(t, fs.Contains(""))
}

func TestFlagSet_Len(t *testing.T) {
	require.Equal(t, 0, NewFlagSet().Len())

	fs := NewFlagSet("flag1", "flag2")
	require.Equal(t, 2, fs.Len())

	fs = NewFlagSet("flag1", "flag2", "flag3", "FLAG2")
	require.Equal(t, 3, fs.Len())
}

func TestFlagSet_ToSlice(t *testing.T) {
	require.True(t, len(NewFlagSet().ToSlice()) == 0)

	fs := NewFlagSet("flag1", "flag2", "FLAG2", "flag3")
	require.True(t, len(fs.ToSlice()) == 3)

	// Check that we return a hard copy.
	fs = NewFlagSet("flag1", "flag2", "flag3")
	sl := fs.ToSlice()
	require.Equal(t, 3, len(sl))

	// Modify something in the returned slice.
	sl[0] = "flag2"

	// It should not be modified in the original flag set.
	require.Equal(t, "flag1", fs.ToSlice()[0])
}

func TestFlagSet_Equals(t *testing.T) {
	// Empty sets are equal.
	require.True(t, NewFlagSet().Equals(NewFlagSet()))

	// Empty set is not equal to nonempty set.
	require.False(t, NewFlagSet().Equals(NewFlagSet("flag1")))

	fs := NewFlagSet("flag1", "flag2", "flag3")
	require.True(t, fs.Equals(fs))
	require.True(t, fs.Equals(NewFlagSet("flag3", "flag2", "flag1")))
	require.True(t, fs.Equals(NewFlagSet("FLAG3", "FLAG2", "FLAG1")))
	require.False(t, fs.Equals(NewFlagSet("flag3", "flag2")))
}

func TestFlagSet_Add(t *testing.T) {
	var (
		fs  = NewFlagSet()
		fs1 = NewFlagSet("flag1")
	)

	fs = fs.Add("flag1")
	require.True(t, fs.Equals(fs1))

	fs = fs.Add()
	require.True(t, fs.Equals(NewFlagSet("flag1")))

	fs = fs.Add("flag1")
	require.True(t, fs.Equals(NewFlagSet("flag1")))

	fs = fs.Add("FLAG1")
	require.ElementsMatch(t, fs.ToSlice(), []string{"flag1"})

	fs = fs.Add("flag2", "flag3")
	require.ElementsMatch(t, fs.ToSlice(), []string{"flag3", "flag2", "flag1"})

	fs = fs.AddFlagSet(NewFlagSet("flag4", "FLAG3"))
	require.ElementsMatch(t, fs.ToSlice(), []string{"flag1", "flag2", "flag3", "flag4"})
}

func TestFlagSet_Toggle(t *testing.T) {
	fs := NewFlagSet("flag1", "flag2")
	fs = fs.Set("flag1", true)
	require.True(t, fs.Equals(NewFlagSet("flag1", "flag2")))

	fs = fs.Set("flag1", false)
	require.True(t, fs.Equals(NewFlagSet("flag2")))

	fs = fs.Set("flag3", false)
	require.True(t, fs.Equals(NewFlagSet("flag2")))

	fs = fs.Set("flag3", true)
	require.True(t, fs.Equals(NewFlagSet("flag2", "flag3")))

	fs = fs.Set("flag3", false)
	require.True(t, fs.Equals(NewFlagSet("flag2")))
}

func TestFlagSet_Remove(t *testing.T) {
	fs := NewFlagSet("flag1", "flag2", "flag3", "flag4")

	fs = fs.Remove()
	require.ElementsMatch(t, fs.ToSlice(), []string{"flag1", "flag2", "flag3", "flag4"})

	fs = fs.Remove("flag4")
	require.ElementsMatch(t, fs.ToSlice(), []string{"flag3", "flag2", "flag1"})

	fs = fs.RemoveFlagSet(NewFlagSet("FLAG3", "flag4"))
	require.ElementsMatch(t, fs.ToSlice(), []string{"flag1", "flag2"})
}
