package state

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
)

func (state *State) GetQuota(ctx context.Context, rootName string) (*imap.QuotaRoot, error) {
	return state.user.GetRemote().GetQuota(ctx, rootName)
}

func (state *State) GetQuotaRoot(ctx context.Context, mailboxName string) ([]string, []*imap.QuotaRoot, error) {
	return state.user.GetRemote().GetQuotaRoot(ctx, mailboxName)
}
