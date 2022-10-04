package response

func Merge(input []Response) []Response {
	if input == nil {
		return nil
	}

	if len(input) < 2 {
		return input
	}

	merged := make([]Response, 0, len(input))

	for _, resp := range input {
		merged = appendOrMergeResponse(merged, resp)
	}

	return merged
}

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
