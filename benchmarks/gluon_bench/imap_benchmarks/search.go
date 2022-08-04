package imap_benchmarks

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/benchmark"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

var searchCountFlag = flag.Uint("search-count", 0, "Total number of messages to search during search benchmarks.")
var searchTextListFlag = flag.String("search-text-list", "", "Use a list of new line separate search queries instead instead of the default list.")
var searchSinceListFlag = flag.String("search-since-list", "", "Use a list of new line dates instead of random generated.")

type SearchText struct {
	queries     []string
	searchCount uint32
}

func NewSearchText() benchmark.Benchmark {
	return NewIMAPBenchmarkRunner(&SearchText{})
}

func (*SearchText) Name() string {
	return "search-text"
}

func (s *SearchText) Setup(ctx context.Context, addr net.Addr) error {
	cl, err := NewClient(addr.String())
	if err != nil {
		return err
	}

	defer CloseClient(cl)

	if err := FillBenchmarkSourceMailbox(cl); err != nil {
		return err
	}

	status, err := cl.Status(*flags.Mailbox, []imap.StatusItem{imap.StatusMessages})
	if err != nil {
		return err
	}

	messageCount := status.Messages

	if messageCount == 0 {
		return fmt.Errorf("mailbox '%v' has no messages", *flags.Mailbox)
	}

	searchCount := uint32(*searchCountFlag)
	if searchCount == 0 {
		searchCount = messageCount / 2
	}

	s.searchCount = searchCount

	if len(*searchTextListFlag) != 0 {
		queries, err := utils.ReadLinesFromFile(*searchTextListFlag)
		if err != nil {
			return err
		}

		s.queries = queries
	} else {
		s.queries = strings.Split(utils.MessageEmbedded, " ")
	}

	return nil
}

func (*SearchText) TearDown(ctx context.Context, addr net.Addr) error {
	return nil
}

func (s *SearchText) Run(ctx context.Context, addr net.Addr) error {
	RunParallelClients(addr, true, func(cl *client.Client, index uint) {
		for i := uint32(0); i < s.searchCount; i++ {
			keywordIndex := rand.Intn(len(s.queries))
			criteria := imap.NewSearchCriteria()

			fieldsStr := []string{"TEXT", s.queries[keywordIndex]}

			fields := xslices.Map(fieldsStr, func(v string) interface{} {
				return interface{}(v)
			})

			if err := criteria.ParseWithCharset(fields, nil); err != nil {
				panic(err)
			}

			if _, err := cl.Search(criteria); err != nil {
				panic(err)
			}
		}
	})

	return nil
}

type SearchSince struct {
	dates []string
}

func NewSearchSince() benchmark.Benchmark {
	return NewIMAPBenchmarkRunner(&SearchSince{})
}

func (*SearchSince) Name() string {
	return "search-since"
}

func (s *SearchSince) Setup(ctx context.Context, addr net.Addr) error {
	cl, err := NewClient(addr.String())
	if err != nil {
		return err
	}

	defer CloseClient(cl)

	if err := FillBenchmarkSourceMailbox(cl); err != nil {
		return err
	}

	status, err := cl.Status(*flags.Mailbox, []imap.StatusItem{imap.StatusMessages})
	if err != nil {
		return err
	}

	messageCount := status.Messages

	if messageCount == 0 {
		return fmt.Errorf("mailbox '%v' has no messages", *flags.Mailbox)
	}

	searchCount := uint32(*searchCountFlag)
	if searchCount == 0 {
		searchCount = messageCount / 2
	}

	if len(*searchSinceListFlag) != 0 {
		dates, err := utils.ReadLinesFromFile(*searchSinceListFlag)
		if err != nil {
			return err
		}

		// validate date formats
		for _, v := range dates {
			if _, err := time.Parse("_2-Jan-2006", v); err != nil {
				return fmt.Errorf("invalid date format in list file: %v", v)
			}
		}

		s.dates = dates
	} else {
		s.dates = make([]string, 0, searchCount)

		for i := uint32(0); i < searchCount; i++ {
			t := time.Date(1980+rand.Intn(40), time.Month(rand.Intn(12)), rand.Intn(28), 0, 0, 0, 0, time.UTC)
			s.dates = append(s.dates, t.Format("02-Jan-2006"))
		}
	}

	return nil
}

func (*SearchSince) TearDown(ctx context.Context, addr net.Addr) error {
	return nil
}

func (s *SearchSince) Run(ctx context.Context, addr net.Addr) error {
	RunParallelClients(addr, true, func(cl *client.Client, index uint) {
		for _, d := range s.dates {
			criteria := imap.NewSearchCriteria()

			fieldsStr := []string{"SINCE", d}

			fields := xslices.Map(fieldsStr, func(v string) interface{} {
				return interface{}(v)
			})

			if err := criteria.ParseWithCharset(fields, nil); err != nil {
				panic(err)
			}

			if _, err := cl.Search(criteria); err != nil {
				panic(err)
			}
		}
	})

	return nil
}
