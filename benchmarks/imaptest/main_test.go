package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/execabs"
	"gopkg.in/yaml.v3"
)

const (
	allowParallel     = false
	doFullIMAPtestLog = false
	gluonLogLevel     = "warn"
	ignoreRecentError = true
)

var (
	rxImapTestError       = regexp.MustCompile(`(?m)^Error: .*$`)
	rxImapTestRecentError = regexp.MustCompile(`Error: .*: Message UID=\d+ has \\Recent flag in multiple sessions:`)
)

func TestIMAPTest(t *testing.T) {
	if path, err := execabs.LookPath("imaptest"); err != nil || path == "" {
		t.Skip("imaptest is not installed")
	}

	r := require.New(t)

	c, err := newConfig("./benchmark.yml")
	r.NoError(err)

	scenarios, err := c.generateScenarios()
	r.NoError(err)

	for _, scenario := range scenarios {
		t.Run(scenario.name, scenario.test)
	}
}

// nolint: musttag
type config struct {
	Cases    []caseConfig
	Settings map[string]settingsConfig
}

type caseConfig struct {
	Users, Clients int
	Settings       []string
	allowFail      bool
}

type settingsConfig map[string]string

func newConfig(path string) (*config, error) {
	rawYAML, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	bm := &config{}

	if err := yaml.Unmarshal(rawYAML, bm); err != nil {
		return nil, err
	}

	return bm, nil
}

func (conf *config) generateScenarios() ([]*scenario, error) {
	var scenarios []*scenario

	i := 0

	for _, c := range conf.Cases {
		for _, settingName := range c.Settings {
			s, ok := conf.Settings[settingName]
			if !ok {
				return nil, fmt.Errorf("wrong config format: cannot find definition for %q setting", settingName)
			}

			sc, err := newScenario(c, settingName, s, 10143+i)
			if err != nil {
				return nil, err
			}

			scenarios = append(scenarios, sc)
			i++
		}
	}

	return scenarios, nil
}

type scenario struct {
	allowFail      bool
	port           string
	users          int
	name           string
	imaptestParams []string
	timeout        time.Duration

	t *testing.T

	ctx context.Context
}

func newScenario(c caseConfig, settingName string, s settingsConfig, port int) (*scenario, error) {
	sc := &scenario{
		allowFail: c.allowFail,
		port:      fmt.Sprintf("%d", port),
		users:     c.Users,
		name:      fmt.Sprintf("u%d_c%d_%s", c.Users, c.Clients, settingName),
		timeout:   time.Duration(time.Second),
	}

	if secs, err := strconv.Atoi(s["secs"]); err == nil {
		sc.timeout = time.Duration(secs) * 2 * time.Second
	}

	// common arguments
	sc.imaptestParams = []string{
		"host=127.0.0.1",
		fmt.Sprintf("port=%d", port),
		"user=user%d@example.com",
		fmt.Sprintf("users=%d", c.Users),
		"pass=pass",
		fmt.Sprintf("clients=%d", c.Clients),
	}

	// scenario specific arguments
	for k, val := range s {
		if val == "false" {
			continue
		}

		if val == "true" {
			sc.imaptestParams = append(sc.imaptestParams, k)
			continue
		}

		sc.imaptestParams = append(sc.imaptestParams, fmt.Sprintf("%s=%s", k, val))
	}

	return sc, nil
}

func (s *scenario) test(t *testing.T) {
	s.t = t

	if allowParallel {
		t.Parallel()
	}

	var cancel context.CancelFunc
	s.ctx, cancel = context.WithTimeout(context.Background(), s.timeout)

	wg := sync.WaitGroup{}
	wg.Add(1)

	defer func() {
		cancel()
		wg.Wait() // to printout log
	}()

	go func() {
		s.runGluon()
		cancel()
		wg.Done()
	}()

	// wait for gluon demo to setup users
	time.Sleep(time.Second)

	s.runIMAPTest()
}

func (s *scenario) runGluon() {
	cmd := execabs.CommandContext(s.ctx, "./gluon-demo")
	cmd.Dir = "../.."
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("GLUON_DIR=%s", s.t.TempDir()),
		fmt.Sprintf("GLUON_USER_COUNT=%d", s.users),
		fmt.Sprintf("GLUON_HOST=127.0.0.1:%s", s.port),
		"GLUON_LOG_LEVEL="+gluonLogLevel,
	)

	out := bytes.NewBuffer(nil)
	cmd.Stderr = out
	cmd.Stdout = out

	err := cmd.Run()

	fmt.Printf("Gluon[%s]:\n%s\nGluonEnd[%s]\n", s.name, out.String(), s.name)

	assert.Error(s.t, err)
	assert.Equal(s.t, "signal: killed", err.Error())
}

func (s *scenario) runIMAPTest() {
	logPath := ""
	if doFullIMAPtestLog {
		logPath = s.t.TempDir() + "imaptest.log"
		s.imaptestParams = append(s.imaptestParams, "output="+logPath)
	}

	cmd := execabs.CommandContext(s.ctx, "imaptest", s.imaptestParams...)

	out := bytes.NewBuffer(nil)
	cmd.Stderr = out
	cmd.Stdout = out

	err := cmd.Run()

	fmt.Printf("IMAPTEST[%s]: %q\n%s\nIMAPTESTEND[%s]\n", s.name, s.imaptestParams, out.String(), s.name)

	assert.NoError(s.t, err)
	assert.False(s.t, hasIMAPLogAnError(out.Bytes(), ignoreRecentError), "Error(s) found in imaptest log.")

	if doFullIMAPtestLog {
		log, err := os.ReadFile(logPath)
		assert.NoError(s.t, err)
		fmt.Println("LOG", s.name, "\n", string(log), "\nLOG", s.name, "END")
	}
}

func hasIMAPLogAnError(log []byte, ignoreRecent bool) bool {
	for _, errMatch := range rxImapTestError.FindAll(log, -1) {
		if ignoreRecent && rxImapTestRecentError.Match(errMatch) {
			continue
		}

		return true
	}

	return false
}

func TestHasIMAPTestAnError(t *testing.T) {
	data := []struct {
		given                string
		want, wantWithIgnore bool
	}{
		{
			given: `Error: unknown error
This is not an error
Error: user1@example.com[510]: Message UID=76 has \Recent flag in multiple sessions: 510 and 511: * 76 FETCH (FLAGS (\Answered \Deleted \Draft \Recent))
Logi List Stat Sele Fetc Fet2 Stor Dele Expu Appe Logo 
100%  50%  50% 100% 100% 100%  50% 100% 100% 100% 100% 
                          30%                  5%      
 524  262  253  500    0    0    0    0    0   63   24 500/500
  52   22   17   24  108    0    0    0    0  158   52 500/500
 136   68   60   95   76   63    0    0    0   72  136 500/500
 `,
			want:           true,
			wantWithIgnore: true,
		},
		{
			given: `This is not an error
Logi List Stat Sele Fetc Fet2 Stor Dele Expu Appe Logo 
100%  50%  50% 100% 100% 100%  50% 100% 100% 100% 100% 
                          30%                  5%      
 524  262  253  500    0    0    0    0    0   63   24 500/500
  52   22   17   24  108    0    0    0    0  158   52 500/500
 136   68   60   95   76   63    0    0    0   72  136 500/500
 `,
			want:           false,
			wantWithIgnore: false,
		},
		{
			given: `This is not an error
Error: user1@example.com[510]: Message UID=76 has \Recent flag in multiple sessions: 510 and 511: * 76 FETCH (FLAGS (\Answered \Deleted \Draft \Recent))
Logi List Stat Sele Fetc Fet2 Stor Dele Expu Appe Logo 
100%  50%  50% 100% 100% 100%  50% 100% 100% 100% 100% 
                          30%                  5%      
 524  262  253  500    0    0    0    0    0   63   24 500/500
  52   22   17   24  108    0    0    0    0  158   52 500/500
 136   68   60   95   76   63    0    0    0   72  136 500/500
 `,
			want:           true,
			wantWithIgnore: false,
		},
	}

	for i, td := range data {
		assert.Equal(t, td.want, hasIMAPLogAnError([]byte(td.given), false), "While not ignoring case %d", i)
		assert.Equal(t, td.wantWithIgnore, hasIMAPLogAnError([]byte(td.given), true), "While ignoring case %d", i)
	}
}
