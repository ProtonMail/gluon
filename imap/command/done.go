package command

import (
	"fmt"
)

type Done struct{}

func (l Done) String() string {
	return fmt.Sprintf("DONE")
}

func (l Done) SanitizedString() string {
	return l.String()
}
