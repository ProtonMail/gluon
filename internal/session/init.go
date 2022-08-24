package session

import (
	"os"
	"strconv"

	"github.com/ProtonMail/gluon/constants"
)

const (
	logIMAPNameLimit           = "GLUON_LOG_IMAP_NAME_LIMIT"
	logIMAPLineLimit           = "GLUON_LOG_IMAP_LINE_LIMIT"
	responseChannelBufferCount = "GLUON_RESPONSE_CHANNEL_BUFFER_COUNT"
)

var (
	maxNameLength      = 16
	maxLineLength      = 120
	channelBufferCount = constants.ChannelBufferCount
)

func init() {
	if val, ok := os.LookupEnv(logIMAPNameLimit); ok {
		valNum, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		maxNameLength = valNum
	}

	if val, ok := os.LookupEnv(logIMAPNameLimit); ok {
		valNum, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		maxNameLength = valNum
	}

	if val, ok := os.LookupEnv(responseChannelBufferCount); ok {
		valNum, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		channelBufferCount = valNum
	}
}
