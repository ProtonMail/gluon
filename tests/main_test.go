package tests

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("GOMSRV_TEST_HOST", "127.0.0.1:2143")
	os.Setenv("GOMSRV_TEST_USER", "dummy@proton.ch")
	os.Setenv("GOMSRV_TEST_PASS", "password")

	os.Exit(m.Run())
}
