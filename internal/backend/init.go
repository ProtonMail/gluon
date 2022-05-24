package backend

import (
	"os"
	"strconv"
)

const fetchConcurrency = "GOMSRV_FETCH_CONCURRENCY"

func init() {
	if val, ok := os.LookupEnv(fetchConcurrency); ok {
		valNum, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}

		maxFetchConcurrency = valNum
	}
}
