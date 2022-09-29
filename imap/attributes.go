package imap

const (
	AttrNoSelect    = `\Noselect`
	AttrNoInferiors = `\Noinferiors`
	AttrMarked      = `\Marked`
	AttrUnmarked    = `\Unmarked`

	// Special Use attributes as defined in RFC-6154.
	AttrAll     = `\All`
	AttrArchive = `\Archive`
	AttrDrafts  = `\Drafts`
	AttrFlagged = `\Flagged`
	AttrJunk    = `\Junk`
	AttrSent    = `\Sent`
	AttrTrash   = `\Trash`
)
