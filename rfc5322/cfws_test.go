package rfc5322

import (
	"strings"
	"testing"

	"github.com/ProtonMail/gluon/internal/unleash"
	"github.com/ProtonMail/gluon/internal/unleash/featureflags"
	"github.com/stretchr/testify/require"
)

func TestParseFWS(t *testing.T) {
	inputs := []string{
		" \t ",
		"\r\n\t",
		" \r\n\t",
		"  \r\n  \r\n  \r\n\t",
		" \t\r\n    ",
	}

	for _, i := range inputs {
		p := newTestRFCParser(i)
		err := parseFWS(p)
		require.NoError(t, err)
	}
}

func TestParserComment(t *testing.T) {
	inputs := []string{
		"(my comment here)",
		"(my comment here )",
		"( my comment here)",
		"( my comment here )",
		"(my\r\n comment here)",
		"(my\r\n (comment) here)",
		"(\\my\r\n (comment) here)",
		"(" + string([]byte{0x7F, 0x8}) + ")",
	}

	for _, i := range inputs {
		p := newTestRFCParser(i)
		err := parseComment(p, 0)
		require.NoError(t, err)
	}
}

// deeplyNestedComment builds a comment nested past the depth limit of 64.
func deeplyNestedComment() string {
	return strings.Repeat("(", 70) + strings.Repeat(")", 70)
}

func TestParserCommentDepthExceeded_KillSwitch_Disabled(t *testing.T) {
	unleash.Init(unleash.NewMockFeatureFlagValueProvider(map[string]bool{
		featureflags.MaximumRFC5322CommentDepthDisabled: false,
	}))
	t.Cleanup(func() { unleash.Init(&unleash.NullFeatureFlagProvider{}) })

	_, err := ParseAddressList(deeplyNestedComment() + "foo@bar.com")
	require.ErrorContains(t, err, "comment nesting depth")
}

func TestParserCommentDepthExceeded_KillSwitch_Enabled(t *testing.T) {
	unleash.Init(unleash.NewMockFeatureFlagValueProvider(map[string]bool{
		featureflags.MaximumRFC5322CommentDepthDisabled: true,
	}))
	t.Cleanup(func() { unleash.Init(&unleash.NullFeatureFlagProvider{}) })

	// With the limit disabled the deep comment is parsed as before (no error from the depth guard).
	_, err := ParseAddressList(deeplyNestedComment() + "foo@bar.com")
	require.NoError(t, err)
}

func TestParserCommentDepthWithinLimit(t *testing.T) {
	unleash.Init(unleash.NewMockFeatureFlagValueProvider(map[string]bool{
		featureflags.MaximumRFC5322CommentDepthDisabled: false,
	}))
	t.Cleanup(func() { unleash.Init(&unleash.NullFeatureFlagProvider{}) })

	addrs, err := ParseAddressList(strings.Repeat("(", 8) + strings.Repeat(")", 8) + "foo@bar.com")
	require.NoError(t, err)
	require.Len(t, addrs, 1)
}

func TestParserCFWS(t *testing.T) {
	inputs := []string{
		" ",
		"(my comment here)",
		" (my comment here) ",
		" \r\n (my comment here)  ",
		" \r\n \r\n (my comment here) \r\n ",
	}

	for _, i := range inputs {
		p := newTestRFCParser(i)
		err := parseCFWS(p)
		require.NoError(t, err)
	}
}
