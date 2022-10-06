package backend

import "errors"

var (
	ErrNoSuchUser   = errors.New("no such user")
	ErrLoginBlocked = errors.New("too many login attempts")
)
