package gluon

import (
	"crypto/tls"
	"io"

	"github.com/ProtonMail/gluon/internal"
)

// Option represents a type that can be used to configure the server.
type Option interface {
	config(server *Server)
}

// WithDelimiter instructs the server to use the given path delimiter instead of the default ('/').
func WithDelimiter(delimiter string) Option {
	return &withDelimiter{
		delimiter: delimiter,
	}
}

type withDelimiter struct {
	delimiter string
}

func (opt withDelimiter) config(server *Server) {
	server.backend.SetDelimiter(opt.delimiter)
}

// WithTLS instructs the server to use the given TLS config.
func WithTLS(cfg *tls.Config) Option {
	return &withTLS{
		cfg: cfg,
	}
}

type withTLS struct {
	cfg *tls.Config
}

func (opt withTLS) config(server *Server) {
	server.tlsConfig = opt.cfg
}

// WithLogger instructs the server to write incoming and outgoing IMAP communication to the given io.Writers.
func WithLogger(in, out io.Writer) Option {
	return &withLogger{
		in:  in,
		out: out,
	}
}

type withLogger struct {
	in, out io.Writer
}

func (opt withLogger) config(server *Server) {
	server.inLogger = opt.in
	server.outLogger = opt.out
}

type withVersionInfo struct {
	versionInfo internal.VersionInfo
}

func (vi *withVersionInfo) config(server *Server) {
	server.versionInfo = vi.versionInfo
}

func WithVersionInfo(vmajor, vminor, vpatch int, name, vendor, supportURL string) Option {
	return &withVersionInfo{
		versionInfo: internal.VersionInfo{
			Name: name,
			Version: internal.Version{
				Major: vmajor,
				Minor: vminor,
				Patch: vpatch,
			},
			Vendor:     vendor,
			SupportURL: supportURL,
		},
	}
}
