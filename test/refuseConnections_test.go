package test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	wwr "github.com/qbeon/webwire-go"
	wwrclt "github.com/qbeon/webwire-go/client"
)

// TestRefuseConnections tests refusal of connection before their upgrade to a
// websocket connection
func TestRefuseConnections(t *testing.T) {
	numClients := 5

	// Initialize server
	server := setupServer(
		t,
		&serverImpl{
			beforeUpgrade: func(
				_ http.ResponseWriter,
				_ *http.Request,
			) wwr.ConnectionOptions {
				// Refuse connections
				return wwr.RefuseConnection("sample reason")
			},
			onRequest: func(
				_ context.Context,
				_ wwr.Connection,
				msg wwr.Message,
			) (wwr.Payload, error) {
				// Expect the following request to not even arrive
				t.Error("Not expected but reached")
				return nil, nil
			},
		},
		wwr.ServerOptions{},
	)
	serverAddr := server.Addr().String()

	clients := make([]*callbackPoweredClient, numClients)
	for i := 0; i < numClients; i++ {
		clt := newCallbackPoweredClient(
			serverAddr,
			wwrclt.Options{
				DefaultRequestTimeout: 2 * time.Second,
				Autoconnect:           wwr.Disabled,
			},
			callbackPoweredClientHooks{},
		)
		defer clt.connection.Close()
		clients[i] = clt

		// Try connect
		require.Error(t, clt.connection.Connect())
	}

	// Try sending requests
	for i := 0; i < numClients; i++ {
		clt := clients[i]
		_, err := clt.connection.Request(context.Background(), "q", nil)
		require.Error(t, err)
		require.IsType(t, wwr.DisconnectedErr{}, err)
	}
}
