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

type matchMailbox struct {
	Name       string
	Subscribed bool
	// EntMBox should be set to nil if there is no such value.
	EntMBox *ent.Mailbox
}

func getMatches(
	ctx context.Context,
	client *ent.Client,
	allMailboxes []matchMailbox,
	ref, pattern, delimiter string,
	subscribed bool,
) (map[string]Match, error) {
	matches := make(map[string]Match)

	mailboxes := make(map[string]matchMailbox)
	for _, mbox := range allMailboxes {
		mailboxes[mbox.Name] = mbox
	}

	for mboxName := range mailboxes {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		default: // fallthrough
		}

		for _, superior := range append(listSuperiors(mboxName, delimiter), mboxName) {
			matchedName, isMatched := match(ref, pattern, delimiter, superior)
			if !isMatched {
				continue
			}

			if _, alreadyMatched := matches[matchedName]; alreadyMatched {
				continue
			}

			mbox, mailboxExists := mailboxes[matchedName]

			match, isMatch, err := prepareMatch(
				ctx, client, matchedName, &mbox,
				pattern, delimiter,
				mboxName == matchedName, mailboxExists, subscribed,
			)
			if err != nil {
				return nil, err
			}

			if isMatch {
				matches[match.Name] = match
			}
		}
	}

	return matches, nil
}

func prepareMatch(
	ctx context.Context,
	client *ent.Client,
	matchedName string,
	mbox *matchMailbox,
	pattern, delimiter string,
	isNotSuperior, mailboxExists, onlySubscribed bool,
) (Match, bool, error) {
	// not match when:
	if onlySubscribed && (mbox == nil || !mbox.Subscribed) && // should be subscribed and it's not
		(isNotSuperior || !strings.HasSuffix(pattern, "%")) { // is not superior or percent wildcard not used
		return Match{}, false, nil
	}

	// add match as NoSelect when:
	if !mailboxExists || // is deleted superior
		matchedName == "" || // is empty request for delimiter response
		onlySubscribed && !mbox.Subscribed { // is unsubscribed superior
		return Match{
			Name:      matchedName,
			Delimiter: delimiter,
			Atts:      imap.NewFlagSet(imap.AttrNoSelect),
		}, true, nil
	}

	var (
		atts imap.FlagSet
	)

	if mbox.EntMBox != nil {
		atts = imap.NewFlagSetFromSlice(xslices.Map(
			mbox.EntMBox.Edges.Attributes,
			func(flag *ent.MailboxAttr) string {
				return flag.Value
			},
		))

		recent, err := db.GetMailboxRecentCount(ctx, client, mbox.EntMBox)
		if err != nil {
			return Match{}, false, err
		}

		if recent > 0 {
			atts.AddToSelf(imap.AttrMarked)
		} else {
			atts.AddToSelf(imap.AttrUnmarked)
		}
	} else {
		atts = imap.NewFlagSet(imap.AttrNoSelect)
	}

	return Match{
		Name:      mbox.Name,
		Delimiter: delimiter,
		Atts:      atts,
	}, true, nil
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
