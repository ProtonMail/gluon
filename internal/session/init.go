package session

import (
	"os"
	"strconv"
)

const (
	logIMAPLineLimit           = "GLUON_LOG_IMAP_LINE_LIMIT"
	responseChannelBufferCount = "GLUON_RESPONSE_CHANNEL_BUFFER_COUNT"
)

var (
	maxLineLength = 120
)

func init() {
	if val, ok := os.LookupEnv(logIMAPLineLimit); ok {
		valNum, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		maxLineLength = valNum
	}
}
