package response

import (
	"fmt"
	"strings"
)

type itemGmailLabels struct {
	labels []string
}

func ItemGmailLabels(labels []string) *itemGmailLabels {
	return &itemGmailLabels{labels: labels}
}

func (g *itemGmailLabels) String() string {
	if len(g.labels) == 0 {
		return "X-GM-LABELS ()"
	}

	quoted := make([]string, len(g.labels))
	for i, label := range g.labels {
		quoted[i] = fmt.Sprintf("%q", label)
	}

	return fmt.Sprintf("X-GM-LABELS (%v)", strings.Join(quoted, " "))
}
