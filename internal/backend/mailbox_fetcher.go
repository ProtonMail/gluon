package backend

import (
	"context"
	"sync/atomic"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/backend/ent/uid"
	"github.com/bradenaw/juniper/iterator"
	"github.com/bradenaw/juniper/stream"
	"github.com/bradenaw/juniper/xslices"
)

// fetcher implements the stream.Streamer interface.
// It is used to process fetched results in parallel from the database.
// TODO: Is the idx threadsafe?
type fetcher struct {
	snap *snapshot
	mbox *ent.Mailbox

	seq imap.SeqSet
	idx int64
}

func (fetcher *fetcher) Next(ctx context.Context) (stream.Stream[*ent.UID], error) {
	idx := atomic.AddInt64(&fetcher.idx, 1)

	if idx == int64(len(fetcher.seq)) {
		return nil, stream.End
	}

	begin, err := fetcher.snap.getMessageBySeq(fetcher.seq[idx].Begin)
	if err != nil {
		return nil, err
	}

	end, err := fetcher.snap.getMessageBySeq(fetcher.seq[idx].End)
	if err != nil {
		return nil, err
	}

	res, err := fetcher.mbox.QueryUIDs().
		Where(uid.UIDGTE(begin.UID), uid.UIDLTE(end.UID)).
		WithMessage().
		All(ctx)
	if err != nil {
		return nil, err
	}

	return stream.FromIterator(iterator.Filter(iterator.Slice(res), func(res *ent.UID) bool {
		return fetcher.snap.hasMessage(res.Edges.Message.MessageID)
	})), nil
}

func (fetcher *fetcher) Close() {
	atomic.StoreInt64(&fetcher.idx, int64(len(fetcher.seq)))
}

func newUIDStream(snap *snapshot, mbox *ent.Mailbox, msg []*snapMsg) stream.Stream[*ent.UID] {
	return stream.Flatten(stream.Stream[stream.Stream[*ent.UID]](&fetcher{
		snap: snap,
		mbox: mbox,
		seq:  imap.NewSeqSet(xslices.Map(msg, func(msg *snapMsg) int { return msg.Seq })),
		idx:  -1,
	}))
}
