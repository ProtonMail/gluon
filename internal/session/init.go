package session

import (
	"os"
	"strconv"
)

const (
	logIMAPNameLimit = "GOMSRV_LOG_IMAP_NAME_LIMIT"
	logIMAPLineLimit = "GOMSRV_LOG_IMAP_LINE_LIMIT"
)

var (
	maxNameLength = 16
	maxLineLength = 120
)

func init() {
	if val, ok := os.LookupEnv(logIMAPNameLimit); ok {
		valNum, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		maxNameLength = valNum
	}

	if val, ok := os.LookupEnv(logIMAPLineLimit); ok {
		valNum, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		maxLineLength = valNum
	}
}
