package imap_benchmarks

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func NewClient(addr string) (*client.Client, error) {
	client, err := client.Dial(addr)
	if err != nil {
		return nil, err
	}

	if err := client.Login(*flags.UserName, *flags.UserPassword); err != nil {
		return nil, err
	}

	return client, nil
}

func AppendToMailbox(cl *client.Client, mailboxName string, literal string, time time.Time, flags ...string) error {
	return cl.Append(mailboxName, flags, time, strings.NewReader(literal))
}

func CloseClient(cl *client.Client) {
	if err := cl.Logout(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to close client: %v\n", err)
	}
}

func FetchMessage(cl *client.Client, sequenceSet *imap.SeqSet, items ...imap.FetchItem) error {
	ch := make(chan *imap.Message)

	go func() {
		for {
			if _, ok := <-ch; !ok {
				break
			}
		}
	}()

	return cl.Fetch(sequenceSet, items, ch)
}

func flagsToInterface(flags ...string) []interface{} {
	return xslices.Map(flags, func(f string) interface{} {
		return interface{}(f)
	})
}

func Store(cl *client.Client, sequenceSet *imap.SeqSet, item string, silent bool, flags ...string) error {
	if !silent {
		ch := make(chan *imap.Message)

		go func() {
			for {
				if _, ok := <-ch; !ok {
					break
				}
			}
		}()

		return cl.Store(sequenceSet, imap.StoreItem(item), flagsToInterface(flags...), ch)
	} else {
		return cl.Store(sequenceSet, imap.StoreItem(item), flagsToInterface(flags...), nil)
	}
}

func UIDStore(cl *client.Client, sequenceSet *imap.SeqSet, item string, silent bool, flags ...string) error {
	if !silent {
		ch := make(chan *imap.Message)

		go func() {
			for {
				if _, ok := <-ch; !ok {
					break
				}
			}
		}()

		return cl.UidStore(sequenceSet, imap.StoreItem(item), flagsToInterface(flags...), ch)
	} else {
		return cl.UidStore(sequenceSet, imap.StoreItem(item), flagsToInterface(flags...), nil)
	}
}

func UIDFetchMessage(cl *client.Client, sequenceSet *imap.SeqSet, items ...imap.FetchItem) error {
	ch := make(chan *imap.Message)

	go func() {
		for {
			if _, ok := <-ch; !ok {
				break
			}
		}
	}()

	return cl.UidFetch(sequenceSet, items, ch)
}

func SequenceListFromFile(path string) ([]*imap.SeqSet, error) {
	result := make([]*imap.SeqSet, 0, 64)

	readFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		line := fileScanner.Text()
		seqSet, err := imap.ParseSeqSet(line)

		if err != nil {
			return nil, err
		}

		result = append(result, seqSet)
	}

	return result, nil
}

func NewSequenceSetAll() *imap.SeqSet {
	seq := &imap.SeqSet{}
	seq.AddRange(1, 0)

	return seq
}

func RandomSequenceSetNum(max uint32) *imap.SeqSet {
	var num uint32
	for num == 0 {
		num = rand.Uint32() % max
	}

	r := &imap.SeqSet{}
	r.AddNum(num)

	return r
}

func RandomSequenceSetRange(max uint32) *imap.SeqSet {
	var start uint32
	for start == 0 {
		start = rand.Uint32() % max
	}

	stop := start
	for stop <= start {
		stop = rand.Uint32() % max
	}

	r := &imap.SeqSet{}
	r.AddRange(start, stop)

	return r
}

type MailboxInfo struct {
	Name     string
	ReadOnly bool
}

func RunParallelClients(addr net.Addr, readOnly bool, fn func(*client.Client, uint)) {
	mailboxes := make([]MailboxInfo, *flags.ParallelClients)
	for i := uint(0); i < *flags.ParallelClients; i++ {
		mailboxes[i] = MailboxInfo{Name: *flags.Mailbox, ReadOnly: readOnly}
	}

	RunParallelClientsWithMailboxes(addr, mailboxes, fn)
}

func RunParallelClientsWithMailboxes(addr net.Addr, mailboxes []MailboxInfo, fn func(*client.Client, uint)) {
	if len(mailboxes) != int(*flags.ParallelClients) {
		panic("Mailbox count doesn't match worker count")
	}

	wg := sync.WaitGroup{}

	for i := uint(0); i < *flags.ParallelClients; i++ {
		wg.Add(1)

		go func(index uint) {
			defer wg.Done()

			cl, err := NewClient(addr.String())

			if err != nil {
				panic(err)
			}

			defer CloseClient(cl)

			mbox := mailboxes[index]

			if _, err := cl.Select(mbox.Name, mbox.ReadOnly); err != nil {
				panic(err)
			}

			fn(cl, index)
		}(i)
	}

	wg.Wait()
}

func FillBenchmarkSourceMailbox(cl *client.Client) error {
	if *flags.FillSourceMailbox != 0 {
		if err := BuildMailbox(cl, *flags.Mailbox, int(*flags.FillSourceMailbox)); err != nil {
			return err
		}
	}

	return nil
}
