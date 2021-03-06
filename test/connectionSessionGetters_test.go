package test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	wwr "github.com/qbeon/webwire-go"
	wwrclt "github.com/qbeon/webwire-go/client"
)

// TestConnectionSessionGetters tests the connection session information
// getter methods such as SessionCreation, SessionKey, SessionInfo and Session
func TestConnectionSessionGetters(t *testing.T) {
	// Initialize server
	server := setupServer(
		t,
		&serverImpl{
			onClientConnected: func(conn wwr.Connection) {
				// Before session creation
				assert.Nil(t, conn.Session())
				assert.Equal(t, time.Time{}, conn.SessionCreation())
				assert.Equal(t, "", conn.SessionKey())
				assert.Nil(t, conn.SessionInfo("uid"))
				assert.Nil(t, conn.SessionInfo("some-number"))

				assert.NoError(t, conn.CreateSession(
					&testAuthenticationSessInfo{
						UserIdent:  "clientidentifiergoeshere", // uid
						SomeNumber: 12345,                      // some-number
					},
				))

				// After session creation
				assert.WithinDuration(
					t,
					time.Now(),
					conn.SessionCreation(),
					1*time.Second,
				)
				assert.NotEqual(t, "", conn.SessionKey())
				uid := conn.SessionInfo("uid")
				assert.NotNil(t, uid)
				assert.IsType(t, string(""), uid)

				someNumber := conn.SessionInfo("some-number")
				assert.NotNil(t, someNumber)
				assert.IsType(t, int(0), someNumber)
			},
		},
		wwr.ServerOptions{},
	)

	// Initialize client
	client := newCallbackPoweredClient(
		server.Addr().String(),
		wwrclt.Options{
			DefaultRequestTimeout: 2 * time.Second,
		},
		callbackPoweredClientHooks{},
	)

	require.NoError(t, client.connection.Connect())
}
