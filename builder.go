package gluon

import (
	"crypto/tls"
	"io"
	"net"
	"os"
	"path/filepath"

	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/internal"
	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/session"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/store"
)

type serverBuilder struct {
	dir                string
	delim              string
	tlsConfig          *tls.Config
	inLogger           io.Writer
	outLogger          io.Writer
	versionInfo        internal.VersionInfo
	cmdExecProfBuilder profiling.CmdProfilerBuilder
	storeBuilder       store.Builder
	reporter           reporter.Reporter
}

func newBuilder() (*serverBuilder, error) {
	return &serverBuilder{
		delim:              "/",
		cmdExecProfBuilder: &profiling.NullCmdExecProfilerBuilder{},
		storeBuilder:       &store.BadgerStoreBuilder{},
		reporter:           &reporter.NullReporter{},
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
		listeners:          make(map[net.Listener]struct{}),
		sessions:           make(map[int]*session.Session),
		inLogger:           builder.inLogger,
		outLogger:          builder.outLogger,
		tlsConfig:          builder.tlsConfig,
		watchers:           make(map[chan events.Event]struct{}),
		storeBuilder:       builder.storeBuilder,
		cmdExecProfBuilder: builder.cmdExecProfBuilder,
		versionInfo:        builder.versionInfo,
		reporter:           builder.reporter,
	}, nil
}
