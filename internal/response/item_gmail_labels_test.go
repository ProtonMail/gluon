package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemGmailLabelsEmpty(t *testing.T) {
	assert.Equal(t, "X-GM-LABELS ()", ItemGmailLabels(nil).String())
	assert.Equal(t, "X-GM-LABELS ()", ItemGmailLabels([]string{}).String())
}

func TestItemGmailLabelsSingle(t *testing.T) {
	assert.Equal(t, `X-GM-LABELS ("Paperless")`, ItemGmailLabels([]string{"Paperless"}).String())
}

func TestItemGmailLabelsMultiple(t *testing.T) {
	assert.Equal(
		t,
		`X-GM-LABELS ("Label1" "Label2" "Label3")`,
		ItemGmailLabels([]string{"Label1", "Label2", "Label3"}).String(),
	)
}

func TestItemGmailLabelsWithSpaces(t *testing.T) {
	assert.Equal(
		t,
		`X-GM-LABELS ("Label With Spaces" "Another One")`,
		ItemGmailLabels([]string{"Label With Spaces", "Another One"}).String(),
	)
}

// A label containing a double quote must be escaped so the response stays a
// valid IMAP quoted string.
func TestItemGmailLabelsQuoting(t *testing.T) {
	assert.Equal(
		t,
		`X-GM-LABELS ("a\"b")`,
		ItemGmailLabels([]string{`a"b`}).String(),
	)
}

func TestItemGmailLabelsSystemLabel(t *testing.T) {
	assert.Equal(
		t,
		`X-GM-LABELS ("\\Inbox")`,
		ItemGmailLabels([]string{`\Inbox`}).String(),
	)
}
