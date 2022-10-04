package response

type Item interface {
	String() string
}

type mergeableItem interface {
	mergeWith(other Item) Item
}
