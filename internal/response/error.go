package response

import "errors"

func FromError(err error) (Response, bool) {
	var no *no

	if errors.As(err, &no) {
		return no, true
	}

	var bad *bad

	if errors.As(err, &bad) {
		return bad, true
	}

	return nil, false
}
