package imap

import (
	"strings"

	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const (
	FlagSeen     = `\Seen`
	FlagAnswered = `\Answered`
	FlagFlagged  = `\Flagged`
	FlagDeleted  = `\Deleted`
	FlagDraft    = `\Draft`
	FlagRecent   = `\Recent` // Read-only!.
)

const (
	FlagSeenLowerCase     = `\seen`
	FlagAnsweredLowerCase = `\answered`
	FlagFlaggedLowerCase  = `\flagged`
	FlagDeletedLowerCase  = `\deleted`
	FlagDraftLowerCase    = `\draft`
	FlagRecentLowerCase   = `\recent` // Read-only!.
)

// FlagSet represents a set of IMAP flags. Flags are case-insensitive and no duplicates are allowed.
type FlagSet map[string]string

// NewFlagSet creates a flag set containing the specified flags.
func NewFlagSet(flags ...string) FlagSet {
	fs := make(FlagSet)

	for _, item := range flags {
		fs.add(item)
	}

	return fs
}

// NewFlagSetFromSlice creates a flag set containing the flags from a slice.
func NewFlagSetFromSlice(flags []string) FlagSet {
	return NewFlagSet(flags...)
}

// Len returns the number of flags in the flag set.
func (fs FlagSet) Len() int {
	return len(fs)
}

// ToSlice Returns the list of flags in the set as a sorted string slice. The returned list is a hard copy of the internal
// slice to avoid direct modifications of the FlagSet value that would break the uniqueness and case insensitivity rules.
func (fs FlagSet) ToSlice() []string {
	flags := maps.Values(fs)

	slices.Sort(flags)

	return flags
}

// Contains returns true if and only if the flag is in the set.
func (fs FlagSet) Contains(flag string) bool {
	_, ok := fs[strings.ToLower(flag)]
	return ok
}

// ContainsUnchecked returns true if and only if the flag is in the set. The flag is not converted to lower case. This
// is useful for cases where we need to check flags in bulk.
func (fs FlagSet) ContainsUnchecked(flag string) bool {
	_, ok := fs[flag]
	return ok
}

// ContainsAny returns true if and only if any of the flags are in the set.
func (fs FlagSet) ContainsAny(flags ...string) bool {
	return xslices.IndexFunc(flags, func(f string) bool {
		return fs.Contains(f)
	}) >= 0
}

// ContainsAll returns true if and only if all of the flags are in the set.
func (fs FlagSet) ContainsAll(flags ...string) bool {
	return xslices.IndexFunc(flags, func(f string) bool {
		return !fs.Contains(f)
	}) < 0
}

// Equals returns true if and only if the two sets are equal (same number of elements and each element of fs is also in otherFs).
func (fs FlagSet) Equals(otherFs FlagSet) bool {
	if fs.Len() != otherFs.Len() {
		return false
	}

	for key := range fs {
		if _, ok := otherFs[key]; !ok {
			return false
		}
	}

	return true
}

// Add adds new flags to the flag set. The function returns false iff no flags was actually added because they're already in the set.
// The case of existing elements is preserved.
func (fs FlagSet) Add(flags ...string) FlagSet {
	return fs.clone().add(flags...)
}

func (fs FlagSet) AddFlagSet(set FlagSet) FlagSet {
	return fs.Add(set.ToSlice()...)
}

func (fs FlagSet) add(flags ...string) FlagSet {
	for _, flag := range flags {
		flagLower := strings.ToLower(flag)

		if fs.ContainsUnchecked(flagLower) {
			continue
		}

		fs[flagLower] = flag
	}

	return fs
}

// Set ensures the flagset either contains or does not contain the given flag.
func (fs FlagSet) Set(flag string, on bool) FlagSet {
	if on {
		return fs.Add(flag)
	} else {
		return fs.Remove(flag)
	}
}

// Remove removes a list of flags from the set. The function returns false iif no flag was actually removed.
func (fs FlagSet) Remove(flags ...string) FlagSet {
	return fs.clone().remove(flags...)
}

func (fs FlagSet) RemoveFlagSet(set FlagSet) FlagSet {
	return fs.Remove(set.ToSlice()...)
}

func (fs FlagSet) remove(flags ...string) FlagSet {
	for _, flag := range flags {
		flagLower := strings.ToLower(flag)

		if !fs.ContainsUnchecked(flagLower) {
			continue
		}

		delete(fs, flagLower)
	}

	return fs
}

// clone creates a hard copy of the flag set.
func (fs FlagSet) clone() FlagSet {
	return NewFlagSetFromSlice(fs.ToSlice())
}
