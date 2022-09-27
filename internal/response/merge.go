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

		lastExists, addResponse, skipExists = mergeExistsReponse(resp, lastExists)
		if addResponse != nil {
			filtered = append(filtered, addResponse)
		}

		lastRecent, addResponse, skipRecent = mergeRecentReponse(resp, lastRecent)
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

func mergeExistsReponse(resp, last Response) (newLast, add Response, skip bool) {
	return mergeTypeReponse(resp, last, isExists, existsHasHigherID)
}

func mergeRecentReponse(resp, last Response) (newLast, add Response, skip bool) {
	return mergeTypeReponse(resp, last, isRecent, recentHasHigherID)
}

func mergeTypeReponse(
	resp, last Response,
	isType func(Response) bool,
	isHigherID func(Response, Response) bool,
) (newLast, add Response, skip bool) {
	if isType(resp) {
		if last == nil || isHigherID(resp, last) {
			return resp, nil, true
		}

		return nil, last, false
	} else {
		if isExists(resp) || isRecent(resp) {
			return last, nil, true
		}
	}

	if last != nil {
		return nil, last, false
	}

	return last, nil, false
}
