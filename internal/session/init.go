package session

import (
	"os"
	"strconv"

	"github.com/ProtonMail/gluon/constants"
)

const (
	logIMAPLineLimit           = "GLUON_LOG_IMAP_LINE_LIMIT"
	responseChannelBufferCount = "GLUON_RESPONSE_CHANNEL_BUFFER_COUNT"
)

var (
	maxLineLength      = 120
	channelBufferCount = constants.ChannelBufferCount
)

func init() {
	if val, ok := os.LookupEnv(logIMAPLineLimit); ok {
		valNum, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		maxLineLength = valNum
	}

	if val, ok := os.LookupEnv(responseChannelBufferCount); ok {
		valNum, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		channelBufferCount = valNum
	}
}
