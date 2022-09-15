package events

type EventUserAdded struct {
	eventBase

	UserID string
}

type EventUserRemoved struct {
	eventBase

	UserID string
}
