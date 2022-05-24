package response

import "fmt"

type itemAppendUID struct {
	uidValidity, messageUID int
}

func ItemAppendUID(uidValidity, messageUID int) *itemAppendUID {
	return &itemAppendUID{
		uidValidity: uidValidity,
		messageUID:  messageUID,
	}
}

func (c *itemAppendUID) String() string {
	return fmt.Sprintf("APPENDUID %v %v", c.uidValidity, c.messageUID)
}
