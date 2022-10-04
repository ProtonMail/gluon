package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
)

type ConnectorImpl interface {
	Connector() connector.Connector
	Sync(ctx context.Context) error
}

type ConnectorBuilder interface {
	New() (ConnectorImpl, error)
}

type connectorFactory struct {
	connectors map[string]ConnectorBuilder
}

func newConnectorFactory() *connectorFactory {
	return &connectorFactory{
		connectors: make(map[string]ConnectorBuilder),
	}
}

func (c *connectorFactory) register(name string, builder ConnectorBuilder) error {
	if _, ok := c.connectors[name]; ok {
		return fmt.Errorf("connector '%v' already exists", name)
	}

	c.connectors[name] = builder

	return nil
}

func (c *connectorFactory) new(name string) (ConnectorImpl, error) {
	builder, ok := c.connectors[name]
	if !ok {
		return nil, fmt.Errorf("no such connector available: '%v'", name)
	}

	return builder.New()
}

var connectorFactoryInstance = newConnectorFactory()

func RegisterConnector(name string, builder ConnectorBuilder) error {
	return connectorFactoryInstance.register(name, builder)
}

func NewConnector(name string) (ConnectorImpl, error) {
	return connectorFactoryInstance.new(name)
}

type DummyConnectorBuilder struct{}

type DummyConnectorImpl struct {
	dummy *connector.Dummy
}

func (d *DummyConnectorImpl) Connector() connector.Connector {
	return d.dummy
}

func (d *DummyConnectorImpl) Sync(ctx context.Context) error {
	d.dummy.ClearUpdates()

	return d.dummy.Sync(ctx)
}

func (*DummyConnectorBuilder) New() (ConnectorImpl, error) {
	addresses := []string{*flags.UserName}
	connector := connector.NewDummy(
		addresses,
		[]byte(*flags.UserPassword),
		time.Second,
		imap.NewFlagSet(`\Answered`, `\Seen`, `\Flagged`, `\Deleted`),
		imap.NewFlagSet(`\Answered`, `\Seen`, `\Flagged`, `\Deleted`),
		imap.NewFlagSet(),
	)

	return &DummyConnectorImpl{dummy: connector}, nil
}

func init() {
	if err := RegisterConnector("dummy", &DummyConnectorBuilder{}); err != nil {
		panic(err)
	}
}
