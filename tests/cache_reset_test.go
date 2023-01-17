package tests

import (
	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestFetchWhenFileDeletedFromCache(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, s *testSession) {
		// create message
		require.NoError(t, doAppendWithClientFromFile(t, client, "INBOX", "testdata/afternoon-meeting.eml", time.Now()))

		// delete message from cache
		require.NoError(t, os.RemoveAll(s.options.dataDir))

		status, err := client.Select("INBOX", false)
		require.NoError(t, err)
		assert.Equal(t, uint32(1), status.Messages)

		// Load message
		fullMessageBytes, err := os.ReadFile("testdata/afternoon-meeting.eml")
		require.NoError(t, err)
		fullMessage := string(fullMessageBytes)

		newFetchCommand(t, client).withItems(goimap.FetchRFC822).fetch("1").forSeqNum(1, func(validator *validatorBuilder) {
			validator.ignoreFlags()
			validator.wantSectionString(goimap.FetchRFC822, func(t testing.TB, literal string) {
				messageFromSection := skipGLUONHeader(literal)
				require.Equal(t, fullMessage, messageFromSection)
			})
		}).checkAndRequireMessageCount(1)
	})
}

func TestSearchWhenFileDeletedFromCache(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, s *testSession) {
		// create message
		require.NoError(t, doAppendWithClientFromFile(t, client, "INBOX", "testdata/afternoon-meeting.eml", time.Now()))

		// delete message from cache
		require.NoError(t, os.RemoveAll(s.options.dataDir))

		status, err := client.Select("INBOX", false)
		require.NoError(t, err)
		assert.Equal(t, uint32(1), status.Messages)

		searchCriteria := goimap.NewSearchCriteria()
		searchCriteria.Text = append(searchCriteria.Text, "3:30")

		seqs, err := client.Search(searchCriteria)
		require.NoError(t, err)
		require.Equal(t, 1, len(seqs))
		require.Equal(t, uint32(1), seqs[0])

	})
}
