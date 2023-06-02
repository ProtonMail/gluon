package imap

// ShortID return a string containing a short version of the given ID. Use only for debug display.
func ShortID(id string) string {
	const l = 12

	if len(id) < l {
		return id
	}

	return id[0:l] + "..."
}
