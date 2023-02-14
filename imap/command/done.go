package command

import (
	"fmt"
)

type DoneCommand struct{}

func (l DoneCommand) String() string {
	return fmt.Sprintf("DONE")
}

func (l DoneCommand) SanitizedString() string {
	return l.String()
}
