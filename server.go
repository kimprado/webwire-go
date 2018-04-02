package webwire

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

const protocolVersion = "1.2"

// Server represents a headless WebWire server instance,
// where headless means there's no HTTP server that's hosting it
type Server struct {
	impl           ServerImplementation
	sessionManager SessionManager
	sessionKeyGen  SessionKeyGenerator

	// State
	shutdown        bool
	shutdownRdy     chan bool
	currentOps      uint32
	opsLock         sync.Mutex
	clientsLock     *sync.Mutex
	clients         []*Client
	sessionsEnabled bool
	sessionRegistry *sessionRegistry

	// Internals
	connUpgrader ConnUpgrader
	warnLog      *log.Logger
	errorLog     *log.Logger
}

// NewServer creates a new WebWire server instance
func NewServer(implementation ServerImplementation, opts ServerOptions) *Server {
	if implementation == nil {
		panic(fmt.Errorf("A headed webwire server requires a server implementation, got nil"))
	}

	opts.SetDefaults()

	srv := Server{
		impl:           implementation,
		sessionManager: opts.SessionManager,
		sessionKeyGen:  opts.SessionKeyGenerator,

		// State
		shutdown:        false,
		shutdownRdy:     make(chan bool),
		currentOps:      0,
		opsLock:         sync.Mutex{},
		clients:         make([]*Client, 0),
		clientsLock:     &sync.Mutex{},
		sessionsEnabled: opts.SessionsEnabled,
		sessionRegistry: newSessionRegistry(opts.MaxSessionConnections),

		// Internals
		connUpgrader: newConnUpgrader(),
		warnLog: log.New(
			opts.WarnLog,
			"WARNING: ",
			log.Ldate|log.Ltime|log.Lshortfile,
		),
		errorLog: log.New(
			opts.ErrorLog,
			"ERROR: ",
			log.Ldate|log.Ltime|log.Lshortfile,
		),
	}

	return &srv
}

// handleSessionRestore handles session restoration (by session key) requests
// and returns an error if the ongoing connection cannot be proceeded
func (srv *Server) handleSessionRestore(msg *Message) error {
	if !srv.sessionsEnabled {
		msg.fail(SessionsDisabledErr{})
		return nil
	}

	key := string(msg.Payload.Data)

	if srv.sessionRegistry.maxConns > 0 &&
		srv.sessionRegistry.SessionConnections(key)+1 > srv.sessionRegistry.maxConns {
		msg.fail(MaxSessConnsReachedErr{})
		return nil
	}

	session, err := srv.sessionManager.OnSessionLookup(key)
	if err != nil {
		msg.fail(nil)
		return fmt.Errorf("CRITICAL: Session search handler failed: %s", err)
	}
	if session == nil {
		msg.fail(SessNotFoundErr{})
		return nil
	}

	// JSON encode the session
	encodedSession, err := json.Marshal(session)
	if err != nil {
		msg.fail(nil)
		return fmt.Errorf("Couldn't encode session object (%v): %s", session, err)
	}

	msg.Client.setSession(session)
	if okay := srv.sessionRegistry.register(msg.Client); !okay {
		panic(fmt.Errorf("The number of concurrent session connections was unexpectedly exceeded"))
	}

	msg.fulfill(Payload{
		Encoding: EncodingUtf8,
		Data:     encodedSession,
	})

	return nil
}

// handleSessionClosure handles session destruction requests
// and returns an error if the ongoing connection cannot be proceeded
func (srv *Server) handleSessionClosure(msg *Message) error {
	if !srv.sessionsEnabled {
		msg.fail(SessionsDisabledErr{})
		return nil
	}

	if !msg.Client.HasSession() {
		// Send confirmation even though no session was closed
		msg.fulfill(Payload{})
		return nil
	}

	srv.deregisterSession(msg.Client)

	// Synchronize session destruction to the client
	if err := msg.Client.notifySessionClosed(); err != nil {
		msg.fail(nil)
		return fmt.Errorf("CRITICAL: Internal server error, "+
			"couldn't notify client about the session destruction: %s",
			err,
		)
	}

	// Reset the session on the client agent
	msg.Client.setSession(nil)

	// Send confirmation
	msg.fulfill(Payload{})

	return nil
}

// handleSignal handles incoming signals
// and returns an error if the ongoing connection cannot be proceeded
func (srv *Server) handleSignal(msg *Message) {
	srv.opsLock.Lock()
	// Ignore incoming signals during shutdown
	if srv.shutdown {
		srv.opsLock.Unlock()
		return
	}
	srv.currentOps++
	srv.opsLock.Unlock()

	srv.impl.OnSignal(context.WithValue(context.Background(), Msg, *msg))

	// Mark signal as done and shutdown the server if scheduled and no ops are left
	srv.opsLock.Lock()
	srv.currentOps--
	if srv.shutdown && srv.currentOps < 1 {
		close(srv.shutdownRdy)
	}
	srv.opsLock.Unlock()
}

// handleRequest handles incoming requests
// and returns an error if the ongoing connection cannot be proceeded
func (srv *Server) handleRequest(msg *Message) {
	srv.opsLock.Lock()
	// Reject incoming requests during shutdown, return special shutdown error
	if srv.shutdown {
		srv.opsLock.Unlock()
		msg.failDueToShutdown()
		return
	}
	srv.currentOps++
	srv.opsLock.Unlock()

	replyPayload, returnedErr := srv.impl.OnRequest(
		context.WithValue(context.Background(), Msg, *msg),
	)
	switch returnedErr.(type) {
	case nil:
		msg.fulfill(replyPayload)
	case ReqErr:
		msg.fail(returnedErr)
	case *ReqErr:
		msg.fail(returnedErr)
	default:
		srv.errorLog.Printf("Internal error during request handling: %s", returnedErr)
		msg.fail(returnedErr)
	}

	// Mark request as done and shutdown the server if scheduled and no ops are left
	srv.opsLock.Lock()
	srv.currentOps--
	if srv.shutdown && srv.currentOps < 1 {
		close(srv.shutdownRdy)
	}
	srv.opsLock.Unlock()
}

// handleMetadata handles endpoint metadata requests
func (srv *Server) handleMetadata(resp http.ResponseWriter) {
	resp.Header().Set("Content-Type", "application/json")
	resp.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(resp).Encode(struct {
		ProtocolVersion string `json:"protocol-version"`
	}{
		protocolVersion,
	})
}

// handleMessage handles incoming messages
func (srv *Server) handleMessage(msg *Message) error {
	switch msg.msgType {
	case MsgSignalBinary:
		fallthrough
	case MsgSignalUtf8:
		fallthrough
	case MsgSignalUtf16:
		srv.handleSignal(msg)

	case MsgRequestBinary:
		fallthrough
	case MsgRequestUtf8:
		fallthrough
	case MsgRequestUtf16:
		srv.handleRequest(msg)

	case MsgRestoreSession:
		return srv.handleSessionRestore(msg)
	case MsgCloseSession:
		return srv.handleSessionClosure(msg)
	}
	return nil
}

// ServeHTTP will make the server listen for incoming HTTP requests
// eventually trying to upgrade them to WebSocket connections
func (srv *Server) ServeHTTP(
	resp http.ResponseWriter,
	req *http.Request,
) {
	// Reject incoming connections during shutdown, pretend the server is temporarily unavailable
	srv.opsLock.Lock()
	if srv.shutdown {
		srv.opsLock.Unlock()
		http.Error(resp, "Server shutting down", http.StatusServiceUnavailable)
		return
	}
	srv.opsLock.Unlock()

	switch req.Method {
	case "OPTIONS":
		srv.impl.OnOptions(resp)
		return
	case "WEBWIRE":
		srv.handleMetadata(resp)
		return
	}

	if !srv.impl.BeforeUpgrade(resp, req) {
		return
	}

	// Establish connection
	conn, err := srv.connUpgrader.Upgrade(resp, req)
	if err != nil {
		srv.errorLog.Print("Upgrade failed:", err)
		return
	}

	// Register connected client
	newClient := newClientAgent(conn, req.Header.Get("User-Agent"), srv)

	srv.clientsLock.Lock()
	srv.clients = append(srv.clients, newClient)
	srv.clientsLock.Unlock()

	// Call hook on successful connection
	srv.impl.OnClientConnected(newClient)

	for {
		// Await message
		message, err := conn.Read()
		if err != nil {
			if newClient.HasSession() {
				// Decrement number of connections for this clients session
				srv.sessionRegistry.deregister(newClient)
			}

			if err.IsAbnormalCloseErr() {
				srv.warnLog.Printf("Abnormal closure error: %s", err)
			}

			newClient.unlink()
			srv.impl.OnClientDisconnected(newClient)
			return
		}

		// Parse message
		var msg Message
		if err := msg.Parse(message); err != nil {
			srv.errorLog.Println("Failed parsing message:", err)
			break
		}

		// Prepare message
		// Reference the client associated with this message
		msg.Client = newClient

		msg.createReplyCallback(newClient, srv)
		msg.createFailCallback(newClient, srv)

		// Handle message
		if err := srv.handleMessage(&msg); err != nil {
			srv.errorLog.Printf("CRITICAL FAILURE: %s", err)
			break
		}
	}
}

func (srv *Server) deregisterSession(clt *Client) {
	srv.sessionRegistry.deregister(clt)
	if err := srv.sessionManager.OnSessionClosed(clt); err != nil {
		srv.errorLog.Printf("OnSessionClosed hook failed: %s", err)
	}
}

// Shutdown appoints a server shutdown and blocks the calling goroutine until the server
// is gracefully stopped awaiting all currently processed signal and request handlers to return.
// During the shutdown incoming connections are rejected with 503 service unavailable.
// Incoming requests are rejected with an error while incoming signals are just ignored
func (srv *Server) Shutdown() {
	srv.opsLock.Lock()
	srv.shutdown = true
	// Don't block if there's no currently processed operations
	if srv.currentOps < 1 {
		return
	}
	srv.opsLock.Unlock()
	<-srv.shutdownRdy
}

// SessionRegistry returns the public interface of the servers session registry
func (srv *Server) SessionRegistry() SessionRegistry {
	return srv.sessionRegistry
}
