package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/response"
)

func flush(ctx context.Context, mailbox *backend.Mailbox, permitExpunge bool, resCh chan response.Response) error {
	res, err := mailbox.Flush(ctx, permitExpunge)
	if err != nil {
		return err
	}

	for _, res := range res {
		resCh <- res
	}

	return nil
}
