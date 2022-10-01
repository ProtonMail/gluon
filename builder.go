package gluon

import (
	"crypto/tls"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/session"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/ProtonMail/gluon/queue"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/store"
	"github.com/ProtonMail/gluon/version"
)

type serverBuilder struct {
	dir                string
	delim              string
	tlsConfig          *tls.Config
	idleBulkTime       time.Duration
	inLogger           io.Writer
	outLogger          io.Writer
	versionInfo        version.Info
	cmdExecProfBuilder profiling.CmdProfilerBuilder
	storeBuilder       store.Builder
	reporter           reporter.Reporter
}

func newBuilder() (*serverBuilder, error) {
	return &serverBuilder{
		delim:              "/",
		cmdExecProfBuilder: &profiling.NullCmdExecProfilerBuilder{},
		storeBuilder:       &store.OnDiskStoreBuilder{},
		reporter:           &reporter.NullReporter{},
		idleBulkTime:       time.Duration(500 * time.Millisecond),
	}, nil
}

func (builder *serverBuilder) build() (*Server, error) {
	if builder.dir == "" {
		dir, err := os.MkdirTemp("", "gluon-*")
		if err != nil {
			return nil, err
		}

		builder.dir = dir
	}

	if err := os.MkdirAll(builder.dir, 0o700); err != nil {
		return nil, err
	}

	backend, err := backend.New(filepath.Join(builder.dir, "backend"), builder.storeBuilder, builder.delim)
	if err != nil {
		return nil, err
	}

	return &Server{
		dir:                builder.dir,
		backend:            backend,
		sessions:           make(map[int]*session.Session),
		serveErrCh:         queue.NewQueuedChannel[error](1, 1),
		serveDoneCh:        make(chan struct{}),
		inLogger:           builder.inLogger,
		outLogger:          builder.outLogger,
		tlsConfig:          builder.tlsConfig,
		idleBulkTime:       builder.idleBulkTime,
		storeBuilder:       builder.storeBuilder,
		cmdExecProfBuilder: builder.cmdExecProfBuilder,
		versionInfo:        builder.versionInfo,
		reporter:           builder.reporter,
	}, nil
}
