package remote

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
)

func (user *User) newContextWithIMAPID(ctx context.Context, id ConnMetadataID) context.Context {
	user.connMetadataStoreLock.RLock()
	defer user.connMetadataStoreLock.RUnlock()

	if v := user.connMetadataStore.GetValue(id, imap.IMAPIDConnMetadataKey); v != nil {
		switch x := v.(type) {
		case imap.ID:
			ctx = imap.NewContextWithIMAPID(ctx, x)
		}
	}

	return ctx
}
