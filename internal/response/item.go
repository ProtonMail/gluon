package response

type Item interface {
	String(isPrivateByDefault bool) string
}

type mergeableItem interface {
	mergeWith(other Item) Item
}
