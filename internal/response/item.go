package response

type Item interface {
	Strings() (raw string, filtered string)
}

type mergeableItem interface {
	mergeWith(other Item) Item
}
