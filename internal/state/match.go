package state

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/bradenaw/juniper/xslices"
)

type Match struct {
	Name      string
	Delimiter string
	Atts      imap.FlagSet
}

func getMatches(
	ctx context.Context,
	client *ent.Client,
	mailboxes []*ent.Mailbox,
	ref, pattern, delimiter string,
	subscribed bool,
) (map[string]Match, error) {
	matches := make(map[string]Match)

	for _, mailbox := range mailboxes {
		if subscribed && !mailbox.Subscribed {
			continue
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		default: //fallthrough
		}

		if name, ok := match(ref, pattern, delimiter, mailbox.Name); ok {
			if mailbox.Name == name {
				atts := imap.NewFlagSetFromSlice(xslices.Map(mailbox.Edges.Attributes, func(flag *ent.MailboxAttr) string {
					return flag.Value
				}))

				recent, err := db.GetMailboxRecentCount(ctx, client, mailbox)
				if err != nil {
					return nil, err
				}

				if recent > 0 {
					atts = atts.Add(imap.AttrMarked)
				} else {
					atts = atts.Add(imap.AttrUnmarked)
				}

				matches[mailbox.Name] = Match{
					Name:      mailbox.Name,
					Delimiter: delimiter,
					Atts:      atts,
				}
			} else {
				matches[name] = Match{
					Name:      name,
					Delimiter: delimiter,
					Atts:      imap.NewFlagSet(imap.AttrNoSelect),
				}
			}
		}
	}

	return matches, nil
}

// GOMSRV-100: validate this implementation.
func match(ref, pattern, del, mailboxName string) (string, bool) {
	if pattern == "" {
		return matchRoot(ref, del)
	}

	rx := fmt.Sprintf("^%v", regexp.QuoteMeta(canon(ref+pattern, del)))

	// If the "%" wildcard is the last character of a mailbox name argument,
	// matching levels of hierarchy are also returned.
	if !strings.HasSuffix(pattern, "%") {
		rx += "$"
	}

	// The character "*" is a wildcard, and matches zero or more characters at this position.
	rx = strings.ReplaceAll(rx, `\*`, ".*")

	// The character "%" is similar to "*", but it does not match a hierarchy delimiter.
	rx = strings.ReplaceAll(rx, "%", fmt.Sprintf("[^%v]*", del))

	if res := regexp.MustCompile(rx).FindAllString(mailboxName, 1); len(res) > 0 {
		return res[0], true
	}

	return "", false
}

// An empty ("" string) mailbox name argument is a special request to
// return the hierarchy delimiter and the root name of the name given
// in the reference. The value returned as the root MAY be the empty
// string if the reference is non-rooted or is an empty string.
func matchRoot(ref, del string) (string, bool) {
	if !strings.Contains(ref, del) {
		return "", true
	}

	var res string

	if strings.HasPrefix(ref, del) {
		res += del
	}

	split := strings.Split(ref, del)

	if len(split) > 0 {
		res += split[0]
	}

	if res != "" && res != del {
		res += del
	}

	return res, true
}

func canon(name, del string) string {
	return strings.Join(xslices.Map(strings.Split(name, del), func(name string) string {
		if strings.EqualFold(name, imap.Inbox) {
			return imap.Inbox
		}

		return name
	}), del)
}
