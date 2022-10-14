package response

import (
	"fmt"
	"time"
)

type itemInternalDate struct {
	date time.Time
}

const internalDateFormat = "02-Jan-2006 15:04:05 -0700"

func ItemInternalDate(date time.Time) *itemInternalDate {
	return &itemInternalDate{date: date}
}

func (c *itemInternalDate) Strings() (raw string, _ string) {
	raw = fmt.Sprintf("INTERNALDATE \"%v\"", c.date.UTC().Format(internalDateFormat))
	return raw, raw
}
