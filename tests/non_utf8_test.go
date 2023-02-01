package tests

import (
	"testing"
	"unicode/utf8"

	"github.com/ProtonMail/gluon/reporter/mock_reporter"
	"github.com/emersion/go-imap/client"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestSSLConnectionOverStartTLS(t *testing.T) {
	ctrl := gomock.NewController(t)
	reporter := mock_reporter.NewMockReporter(ctrl)

	defer ctrl.Finish()

	// Ensure the nothing is reported when connecting via TLS connection if we are not running with TLS
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withReporter(reporter)), func(_ *client.Client, session *testSession) {
		_, err := client.DialTLS(session.listener.Addr().String(), nil)
		require.Error(t, err)
	})
}

func TestNonUtf8CommandTriggersReporter(t *testing.T) {
	ctrl := gomock.NewController(t)
	reporter := mock_reporter.NewMockReporter(ctrl)

	defer ctrl.Finish()

	reporter.EXPECT().ReportMessageWithContext("Received invalid UTF-8 command", gomock.Any()).Return(nil).Times(1)

	// Ensure the nothing is reported when connecting via TLS connection if we are not running with TLS
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withReporter(reporter)), func(c *testConnection, session *testSession) {
		// Encode "ééé" as ISO-8859-1.
		b := enc("ééé", "ISO-8859-1")

		// Assert that b is no longer valid UTF-8.
		require.False(t, utf8.Valid(b))

		// This will fail and produce a report
		c.Cf(`TAG SEARCH CHARSET ISO-8859-1 BODY ` + string(b))
	})
}
