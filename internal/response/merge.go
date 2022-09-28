package response

func Merge(input []Response) []Response {
	if input == nil {
		return nil
	}

	var lastExists, lastRecent Response

	filtered := []Response{}

	for _, resp := range input {
		var addResponse Response

		var skipExists, skipRecent bool

		lastExists, addResponse, skipExists = mergeExistsResponse(resp, lastExists)
		if addResponse != nil {
			filtered = append(filtered, addResponse)
		}

		lastRecent, addResponse, skipRecent = mergeRecentResponse(resp, lastRecent)
		if addResponse != nil {
			filtered = append(filtered, addResponse)
		}

		if skipRecent && skipExists {
			continue
		}

		filtered = append(filtered, resp)
	}

	if lastExists != nil {
		filtered = append(filtered, lastExists)
	}

	if lastRecent != nil {
		filtered = append(filtered, lastRecent)
	}

	return filtered
}

func mergeExistsResponse(resp, last Response) (newLast, add Response, skip bool) {
	return mergeTypeResponse(resp, last, isExists, existsHasHigherID)
}

func mergeRecentResponse(resp, last Response) (newLast, add Response, skip bool) {
	return mergeTypeResponse(resp, last, isRecent, recentHasHigherID)
}

func mergeTypeResponse(
	resp, last Response,
	isType func(Response) bool,
	isHigherID func(Response, Response) bool,
) (newLast, add Response, skip bool) {
	if isType(resp) {
		if last == nil || isHigherID(resp, last) {
			return resp, nil, true
		}

		panic("response decreased ID for exists or recent without expunge")
	}

	if isExists(resp) || isRecent(resp) {
		return last, nil, true
	}

	if last != nil {
		return nil, last, false
	}

	return last, nil, false
}
