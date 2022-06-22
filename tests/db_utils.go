package tests

import (
	"context"

	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/stretchr/testify/require"
)

func dbCheckUserMessageCount(s *testSession, user string, expectedCount int) {
	err := s.withUserDB(user, func(ent *ent.Client, ctx context.Context) {
		val, err := ent.Message.Query().Count(ctx)
		require.NoError(s.tb, err)
		require.Equal(s.tb, expectedCount, val)
	})
	require.NoError(s.tb, err)
}
