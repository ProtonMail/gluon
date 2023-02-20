package tests

import (
	"testing"

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
