package tests

import (
	"crypto/tls"
	"reflect"
	"testing"

	"github.com/bradenaw/juniper/xslices"
	"github.com/stretchr/testify/require"
)

// nolint:gosec
func _TestNonUTF8(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(_ *testConnection, s *testSession) {
		// Create a new connection.
		c := s.newConnection()

		// Things work fine when the command is valid UTF-8.
		c.C("tag capability").OK("tag")

		// Performing a TLS handshake should fail; the server will drop the connection.
		require.Error(t, tls.Client(c.conn, &tls.Config{InsecureSkipVerify: true}).Handshake())

		// We should have reported the bad UTF-8 command.
		require.True(t, xslices.Any(s.reporter.getReports(), func(report report) bool {
			return reflect.DeepEqual(report.val, "Received invalid UTF-8 command")
		}))
	})
}
