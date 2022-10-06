package response

func Merge(input []Response) []Response {
	if len(input) < 2 {
		return input
	}

	merged := make([]Response, 0, len(input))

	for _, resp := range input {
		merged = appendOrMergeResponse(merged, resp)
	}

	return merged
}

// appendOrMergeResponse will find out if `unmerged` can be merged with some
// element of `input`. It starts last thing first (i.e. the newest update).
//
// There are certain response types which cannot be merged together (e.g. recent and
// exist) but one response can ignore (canSkip == true) the other and try to
// merge with older responses (e.g. exists.canSkip(other) return true for
// recent and fetch).
func appendOrMergeResponse(input []Response, unmerged Response) []Response {
	mergeable, ok := unmerged.(mergeableResponse)
	if !ok {
		return append(input, unmerged)
	}

	wasMerged := false

	for i := len(input) - 1; i >= 0; i-- {
		merged := mergeable.mergeWith(input[i])
		if merged != nil {
			wasMerged = true
			input[i] = merged

			break
		}

		if !mergeable.canSkip(input[i]) {
			break
		}
	}

	if !wasMerged {
		input = append(input, unmerged)
	}

	return input
}

func appendOrMergeItem(input []Item, unmerged Item) []Item {
	mergeable, ok := unmerged.(mergeableItem)
	if !ok {
		return append(input, unmerged)
	}

	wasMerged := false

	for i := range input {
		merged := mergeable.mergeWith(input[i])
		if merged != nil {
			wasMerged = true
			input[i] = merged

			break
		}
	}

	if !wasMerged {
		input = append(input, unmerged)
	}

	return input
}
