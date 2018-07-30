package test

import (
	"context"
	"reflect"
	"testing"
	"time"

	wwr "github.com/qbeon/webwire-go"
	wwrclt "github.com/qbeon/webwire-go/client"
)

// TestClientRequestDisconnected tests sending requests on disconnected clients
func TestClientRequestDisconnected(t *testing.T) {
	// Initialize webwire server given only the request
	server := setupServer(
		t,
		&serverImpl{
			onRequest: func(
				_ context.Context,
				_ wwr.Connection,
				_ wwr.Message,
			) (wwr.Payload, error) {
				return nil, nil
			},
		},
		wwr.ServerOptions{},
	)

	// Initialize client and skip manual connection establishment
	client := newCallbackPoweredClient(
		server.Addr().String(),
		wwrclt.Options{
			DefaultRequestTimeout: 2 * time.Second,
		},
		callbackPoweredClientHooks{},
	)

	// Send request and await reply
	if _, err := client.connection.Request(
		"",
		wwr.NewPayload(wwr.EncodingBinary, []byte("testdata")),
	); err != nil {
		t.Fatalf(
			"Expected request to automatically connect, got error: %s | %s",
			reflect.TypeOf(err),
			err,
		)
	}
}
