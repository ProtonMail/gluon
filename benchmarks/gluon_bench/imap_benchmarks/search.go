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

var (
	searchCountFlag     = flag.Uint("imap-search-count", 0, "Total number of messages to search during search benchmarks.")
	searchTextListFlag  = flag.String("imap-search-text-list", "", "Use a list of new line separate search queries instead instead of the default list.")
	searchSinceListFlag = flag.String("imap-search-since-list", "", "Use a list of new line dates instead of random generated.")
	searchCmdQueryFlag  = flag.String("imap-search-cmd", "", "Search command to execute e.g.: \"OR BEFORE <> SINCE <>\"")
)

type SearchQuery interface {
	Name() string
	Setup(context.Context, *client.Client, uint32) error
	Run(context.Context, *client.Client, uint) error
	TearDown(context.Context, *client.Client) error
}

type Search struct {
	*stateTracker
	query SearchQuery
}

func NewSearch(query SearchQuery) benchmark.Benchmark {
	return NewIMAPBenchmarkRunner(&Search{stateTracker: newStateTracker(), query: query})
}

func (s *Search) Name() string {
	return s.query.Name()
}

func (s *Search) Setup(ctx context.Context, addr net.Addr) error {
	return WithClient(addr, func(cl *client.Client) error {
		if _, err := s.createAndFillRandomMBox(cl); err != nil {
			return err
		}

		searchCount := uint32(*searchCountFlag)
		if searchCount == 0 {
			searchCount = uint32(*flags.IMAPMessageCount)
		}

		if err := s.query.Setup(ctx, cl, searchCount); err != nil {
			return err
		}

		return nil
	})
}

func (s *Search) TearDown(ctx context.Context, addr net.Addr) error {
	return WithClient(addr, func(cl *client.Client) error {
		if err := s.query.TearDown(ctx, cl); err != nil {
			return err
		}

		return s.cleanup(cl)
	})
}

func (s *Search) Run(ctx context.Context, addr net.Addr) error {
	RunParallelClientsWithMailbox(addr, s.MBoxes[0], true, func(cl *client.Client, index uint) {
		if err := s.query.Run(ctx, cl, index); err != nil {
			panic(err)
		}
	})

	return nil
}

type SearchTextQuery struct {
	queries     []string
	searchCount uint32
}

func (s *SearchTextQuery) Name() string {
	return "imap-search-text"
}

func (s *SearchTextQuery) Setup(ctx context.Context, cl *client.Client, searchCount uint32) error {
	if len(*searchTextListFlag) != 0 {
		queries, err := utils.ReadLinesFromFile(*searchTextListFlag)
		if err != nil {
			return err
		}

		s.queries = queries
	} else {
		s.queries = strings.Split(utils.MessageEmbedded, " ")
	}

	s.searchCount = searchCount

	return nil
}

func (s *SearchTextQuery) TearDown(ctx context.Context, cl *client.Client) error {
	return nil
}

func (s *SearchTextQuery) Run(ctx context.Context, cl *client.Client, workerIndex uint) error {
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

	return nil
}

type SearchSinceQuery struct {
	dates []string
}

func (*SearchSinceQuery) Name() string {
	return "imap-search-since"
}

func (s *SearchSinceQuery) Setup(ctx context.Context, cl *client.Client, searchCount uint32) error {
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

func (*SearchSinceQuery) TearDown(ctx context.Context, cl *client.Client) error {
	return nil
}

func (s *SearchSinceQuery) Run(ctx context.Context, cl *client.Client, workerIndex uint) error {
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

	return nil
}

type SearchCmdQuery struct {
	criteria    *imap.SearchCriteria
	searchCount uint32
}

func (*SearchCmdQuery) Name() string {
	return "imap-search-cmd"
}

func (s *SearchCmdQuery) Setup(ctx context.Context, cl *client.Client, searchCount uint32) error {
	s.criteria = imap.NewSearchCriteria()

	if len(*searchCmdQueryFlag) == 0 {
		return fmt.Errorf("please provide a query with -imap-search-cmd")
	}

	queries := strings.Split(*searchCmdQueryFlag, " ")

	fields := xslices.Map(queries, func(v string) interface{} {
		return interface{}(v)
	})

	if err := s.criteria.ParseWithCharset(fields, nil); err != nil {
		return err
	}

	s.searchCount = searchCount

	return nil
}

func (*SearchCmdQuery) TearDown(ctx context.Context, cl *client.Client) error {
	return nil
}

func (s *SearchCmdQuery) Run(ctx context.Context, cl *client.Client, workerIndex uint) error {
	for i := uint32(0); i < s.searchCount; i++ {
		if _, err := cl.Search(s.criteria); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	benchmark.RegisterBenchmark(NewSearch(&SearchSinceQuery{}))
	benchmark.RegisterBenchmark(NewSearch(&SearchTextQuery{}))
	benchmark.RegisterBenchmark(NewSearch(&SearchCmdQuery{}))
}
