package events

type UserAdded struct {
	eventBase

	UserID string
}

type UserRemoved struct {
	eventBase

	UserID string
}
