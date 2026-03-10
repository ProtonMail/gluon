package connectionlimiter

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func imapID(name string) imap.IMAPID {
	return imap.IMAPID{Name: name}
}

func TestTryBind_FirstConnectionAllowed(t *testing.T) {
	limits := NewLimits(map[Client]int{ClientAppleMail: 3}, 1)
	fallbackLimits := NewDefaultLimits()
	useFallback := false

	l := NewConnectionLimiter(limits, fallbackLimits)
	allowed, key, current, max := l.TryBind(1, imapID("MacOS X Mail"), useFallback)
	require.True(t, allowed)
	assert.Equal(t, ClientAppleMail, key)
	assert.Equal(t, 1, current)
	assert.Equal(t, 3, max)
}

func TestTryBind_OverLimitDenied(t *testing.T) {
	limits := NewLimits(map[Client]int{ClientAppleMail: 2}, 1)
	fallbackLimits := NewDefaultLimits()
	useFallback := false

	l := NewConnectionLimiter(limits, fallbackLimits)
	l.TryBind(1, imapID("MacOS X Mail"), useFallback)
	l.TryBind(2, imapID("MacOS X Mail"), useFallback)
	allowed, key, current, max := l.TryBind(3, imapID("MacOS X Mail"), useFallback)
	require.False(t, allowed)
	assert.Equal(t, ClientAppleMail, key)
	assert.Equal(t, 2, current)
	assert.Equal(t, 2, max)
}

func TestUnbind_FreesSlot(t *testing.T) {
	limits := NewLimits(map[Client]int{ClientAppleMail: 2}, 1)
	fallbackLimits := NewDefaultLimits()
	useFallback := false

	l := NewConnectionLimiter(limits, fallbackLimits)
	l.TryBind(1, imapID("MacOS X Mail"), useFallback)
	l.TryBind(2, imapID("MacOS X Mail"), useFallback)
	l.Unbind(1)
	allowed, _, current, _ := l.TryBind(3, imapID("MacOS X Mail"), useFallback)
	require.True(t, allowed)
	assert.Equal(t, 2, current)
}

func TestUnbind_UnknownSessionNoop(t *testing.T) {
	limits := NewLimits(map[Client]int{ClientAppleMail: 2}, 1)
	fallbackLimits := NewDefaultLimits()
	useFallback := false

	l := NewConnectionLimiter(limits, fallbackLimits)
	l.TryBind(1, imapID("MacOS X Mail"), useFallback)
	l.Unbind(999) // never bound
	_, _, current, _ := l.TryBind(2, imapID("MacOS X Mail"), useFallback)
	assert.Equal(t, 2, current)
}

func TestTryBind_SameSessionSameClientNoop(t *testing.T) {
	limits := NewLimits(map[Client]int{ClientAppleMail: 2}, 1)
	fallbackLimits := NewDefaultLimits()
	useFallback := false

	l := NewConnectionLimiter(limits, fallbackLimits)
	l.TryBind(1, imapID("MacOS X Mail"), useFallback)
	allowed, key, current, max := l.TryBind(1, imapID("MacOS X Mail"), useFallback)
	require.True(t, allowed)
	assert.Equal(t, ClientAppleMail, key)
	assert.Equal(t, 1, current)
	assert.Equal(t, 2, max)
}

func TestTryBind_RebindToDifferentClient(t *testing.T) {
	limits := NewLimits(map[Client]int{
		ClientAppleMail: 2,
		ClientOutlook:   2,
	}, 1)
	fallbackLimits := NewDefaultLimits()
	useFallback := false

	l := NewConnectionLimiter(limits, fallbackLimits)
	l.TryBind(1, imapID("MacOS X Mail"), useFallback)
	allowed, key, cur, _ := l.TryBind(1, imapID("Microsoft Outlook"), useFallback)
	require.True(t, allowed)
	assert.Equal(t, ClientOutlook, key)
	assert.Equal(t, 1, cur)
	_, _, appleCur, _ := l.TryBind(2, imapID("MacOS X Mail"), useFallback)
	assert.Equal(t, 1, appleCur)
}

func TestTryBind_UnknownClientUsesUnknownLimit(t *testing.T) {
	limits := NewLimits(map[Client]int{ClientAppleMail: 10}, 2)
	fallbackLimits := NewDefaultLimits()
	useFallback := false

	l := NewConnectionLimiter(limits, fallbackLimits)
	l.TryBind(1, imapID("SomeOtherClient"), useFallback)
	l.TryBind(2, imapID("Unknown"), useFallback)
	allowed, key, current, max := l.TryBind(3, imapID("Custom"), useFallback)
	require.False(t, allowed)
	assert.Equal(t, ClientUnknown, key)
	assert.Equal(t, 2, current)
	assert.Equal(t, 2, max)
}

func TestTryBind_UnlimitedLimit(t *testing.T) {
	limits := NewLimits(map[Client]int{ClientAppleMail: 0}, 1)
	fallbackLimits := NewDefaultLimits()
	useFallback := false

	l := NewConnectionLimiter(limits, fallbackLimits)
	for i := 1; i <= 5; i++ {
		allowed, key, current, max := l.TryBind(i, imapID("MacOS X Mail"), useFallback)
		require.True(t, allowed)
		assert.Equal(t, ClientAppleMail, key)
		assert.Equal(t, i, current)
		assert.Equal(t, 0, max)
	}
}

func TestTryBind_UseFallbackLimits(t *testing.T) {
	limits := NewLimits(map[Client]int{ClientAppleMail: 1}, 1)
	fallbackLimits := NewDefaultLimits()

	useFallback := true

	l := NewConnectionLimiter(limits, fallbackLimits)
	for i := 1; i <= 5; i++ {
		allowed, key, current, max := l.TryBind(i, imapID("Mac OS X Mail"), useFallback)
		require.True(t, allowed)
		assert.Equal(t, ClientAppleMail, key)
		assert.Equal(t, i, current)
		assert.Equal(t, fallbackLimits.PerClient[ClientAppleMail], max)
	}
}

func TestNormalizeClientKey_AppleMail(t *testing.T) {
	assert.Equal(t, ClientAppleMail, normalizeClientKey(imapID("MacOS X Mail")))
}
func TestNormalizeClientKey_OutlookAndThunderbird(t *testing.T) {
	assert.Equal(t, ClientOutlook, normalizeClientKey(imapID("Microsoft Outlook")))
	assert.Equal(t, ClientThunderbird, normalizeClientKey(imapID("Thunderbird")))
}
func TestNormalizeClientKey_Unknown(t *testing.T) {
	assert.Equal(t, ClientUnknown, normalizeClientKey(imapID("")))
	assert.Equal(t, ClientUnknown, normalizeClientKey(imapID("Custom Client")))
}
