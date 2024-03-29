package gluon

import (
	"context"
	"net"
	"testing"

	"github.com/ProtonMail/gluon/events"
	"github.com/stretchr/testify/require"
)

func _TestServer(t *testing.T) {
	server, err := New()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Get an event channel.
	eventCh := server.AddWatcher(events.ListenerAdded{}, events.ListenerRemoved{})

	// Create a listener.
	l, err := net.Listen("tcp", net.JoinHostPort("localhost", "0"))
	require.NoError(t, err)

	// The first listen is successful.
	require.NoError(t, server.Serve(ctx, l))
	require.Equal(t, events.ListenerAdded{Addr: l.Addr()}, <-eventCh)

	// The server closes successfully.
	require.NoError(t, server.Close(ctx))
	require.Equal(t, events.ListenerRemoved{Addr: l.Addr()}, <-eventCh)
}
