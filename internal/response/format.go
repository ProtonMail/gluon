package response

import "strings"

func join(items []string, withDel ...string) string {
	var del string

	if len(withDel) > 0 {
		del = withDel[0]
	} else {
		del = " "
	}

	return strings.Join(items, del)
}
