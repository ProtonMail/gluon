package session

import (
	"errors"

	"github.com/ProtonMail/gluon/imap/command"
)

// errGmailExtensionDisabled is returned when a client issues an X-GM-LABELS
// command while the Gmail X-GM-EXT-1 extension is not enabled. In that case the
// capability is not advertised, so a compliant client will not send these
// commands; this guards against non-compliant clients that do anyway.
var errGmailExtensionDisabled = errors.New("X-GM-EXT-1 extension is not enabled")

// fetchHasGmailLabels reports whether a FETCH command requests the X-GM-LABELS attribute.
func fetchHasGmailLabels(cmd *command.Fetch) bool {
	for _, attr := range cmd.Attributes {
		if _, ok := attr.(*command.FetchAttributeGmailLabels); ok {
			return true
		}
	}

	return false
}

// searchKeysHaveGmailLabels reports whether any (possibly nested) SEARCH key uses X-GM-LABELS.
func searchKeysHaveGmailLabels(keys []command.SearchKey) bool {
	for _, key := range keys {
		if searchKeyHasGmailLabels(key) {
			return true
		}
	}

	return false
}

func searchKeyHasGmailLabels(key command.SearchKey) bool {
	switch k := key.(type) {
	case *command.SearchKeyGmailLabels:
		return true
	case *command.SearchKeyNot:
		return searchKeyHasGmailLabels(k.Key)
	case *command.SearchKeyOr:
		return searchKeyHasGmailLabels(k.Key1) || searchKeyHasGmailLabels(k.Key2)
	case *command.SearchKeyList:
		return searchKeysHaveGmailLabels(k.Keys)
	default:
		return false
	}
}
