package tests

import (
	"context"
	"github.com/ProtonMail/gluon/db"

	"github.com/stretchr/testify/require"
)

func dbCheckUserMessageCount(s *testSession, user string, expectedCount int) {
	err := s.withUserDB(user, func(ent db.Client, ctx context.Context) {
		val, err := db.ClientReadType(ctx, ent, func(ctx context.Context, only db.ReadOnly) (int, error) {
			return only.GetTotalMessageCount(ctx)
		})
		require.NoError(s.tb, err)
		require.Equal(s.tb, expectedCount, val)
	})
	require.NoError(s.tb, err)
}
