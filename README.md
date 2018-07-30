<!-- HEADER -->
<h1 align="center">
	<br>
	<a href="https://github.com/qbeon/webwire-go"><img src="https://cdn.rawgit.com/qbeon/webwire-go/c7c2c74e/docs/img/webwire_logo.svg" alt="WebWire" width="256"></a>
	<br>
	<br>
	WebWire for <a href="https://golang.org/">Go</a>
	<br>
	<sub>An asynchronous duplex messaging library</sub>
</h1>
<p align="center">
	<a href="https://travis-ci.org/qbeon/webwire-go">
		<img src="https://travis-ci.org/qbeon/webwire-go.svg?branch=master" alt="Travis CI: build status">
	</a>
	<a href="https://coveralls.io/github/qbeon/webwire-go?branch=master">
		<img src="https://coveralls.io/repos/github/qbeon/webwire-go/badge.svg?branch=master" alt="Coveralls: Test Coverage">
	</a>
	<a href="https://goreportcard.com/report/github.com/qbeon/webwire-go">
		<img src="https://goreportcard.com/badge/github.com/qbeon/webwire-go" alt="GoReportCard">
	</a>
	<a href="https://codebeat.co/projects/github-com-qbeon-webwire-go-master">
		<img src="https://codebeat.co/badges/809181da-797c-4cdd-bb23-d0324935f3b0" alt="CodeBeat: Status">
	</a>
	<a href="https://codeclimate.com/github/qbeon/webwire-go/maintainability">
		<img src="https://api.codeclimate.com/v1/badges/243a45cacec7d850c64d/maintainability" alt="CodeClimate: Maintainability">
	</a>
	<br>
	<a href="https://opensource.org/licenses/MIT">
		<img src="https://img.shields.io/badge/License-MIT-green.svg" alt="Licence: MIT">
	</a>
	<a href="https://godoc.org/github.com/qbeon/webwire-go">
		<img src="https://godoc.org/github.com/zalando/skipper?status.svg" alt="GoDoc">
	</a>
	<a href="https://www.paypal.me/romshark">
		<img src="https://img.shields.io/badge/Donate-PayPal-green.svg" alt="Donate">
	</a>
</p>
<br>

<!-- CONTENT -->
WebWire is a high-level asynchronous duplex messaging library built on top of [WebSockets](https://developer.mozilla.org/de/docs/WebSockets) and an open source binary message protocol with builtin authentication and support for UTF8 and UTF16 encoding.
The [webwire-go](https://github.com/qbeon/webwire-go) library provides both a client and a server implementation for the Go programming language. An official [JavaScript client](https://github.com/qbeon/webwire-js) implementation is also available. WebWire provides a compact set of useful features that are not available and/or cumbersome to implement on raw WebSockets such as Request-Reply, Sessions, Thread-safety etc.
<br>

#### Table of Contents
- [Installation](#installation)
  - [Dep](#dep)
  - [Go Get](#go-get)
- [Contribution](#contribution)
  - [Maintainers](#maintainers)
- [WebWire Binary Protocol](#webwire-binary-protocol)
- [Examples](#examples)
- [Features](#features)
  - [Request-Reply](#request-reply)
  - [Client-side Signals](#client-side-signals)
  - [Server-side Signals](#server-side-signals)
  - [Namespaces](#namespaces)
  - [Sessions](#sessions)
  - [Automatic Session Restoration](#automatic-session-restoration)
  - [Automatic Connection Maintenance](#automatic-connection-maintenance)
  - [Concurrency](#concurrency)
  - [Thread Safety](#thread-safety)
  - [Hooks](#hooks)
    - [Server-side Hooks](#server-side-hooks)
    - [Client-side Hooks](#client-side-hooks)
  - [Graceful Shutdown](#graceful-shutdown)
  - [Seamless JavaScript Support](#seamless-javascript-support)
- [Dependencies](#dependencies)


## Installation
Choose any stable release from [the available release tags](https://github.com/qbeon/webwire-go/releases) and copy the source code into your project's vendor directory: ```$YOURPROJECT/vendor/github.com/qbeon/webwire-go```. All necessary transitive [dependencies](https://github.com/qbeon/webwire-go#dependencies) are already embedded into the `webwire-go` repository.

### Dep
If you're using [dep](https://github.com/golang/dep), just use [dep ensure](https://golang.github.io/dep/docs/daily-dep.html#adding-a-new-dependency) to add a specific version of webwire-go including all its transitive dependencies to your project: ```dep ensure -add github.com/qbeon/webwire-go@v1.0.0-rc1```. This will remove all embedded transitive dependencies and move them to your projects `vendor` directory.

### Go Get
You can also use [go get](https://golang.org/cmd/go/#hdr-Download_and_install_packages_and_dependencies): ```go get github.com/qbeon/webwire-go``` but beware that this will fetch the latest commit of the [master branch](https://github.com/qbeon/webwire-go/commits/master) which is currently **not yet** considered a stable release branch. It's therefore recommended to use [dep](https://github.com/qbeon/webwire-go#dep) instead.

## Contribution
Feel free to report bugs and propose new features or changes in the [Issues section](https://github.com/qbeon/webwire-go/issues).

### Maintainers
| Maintainer | Role | Specialization |
| :--- | :--- | :--- |
| **[Roman Sharkov](https://github.com/romshark)** | Core Maintainer | Dev (Go, JavaScript) |
| **[Daniil Trishkin](https://github.com/FromZeus)** | CI Maintainer | DevOps |

## WebWire Binary Protocol
WebWire is built for speed and portability implementing an open source binary protocol.
![Protocol Subset Diagram](https://github.com/qbeon/webwire-go/blob/master/docs/img/wwr_msgproto_diagram.svg)

The first byte defines the [type of the message](https://github.com/qbeon/webwire-go/blob/master/message.go#L51). Requests and replies contain an incremental 8-byte identifier that must be unique in the context of the senders' session. A 0 to 255 bytes long 7-bit ASCII encoded name is contained in the header of a signal or request message.
A header-padding byte is applied in case of UTF16 payload encoding to properly align the payload sequence.
Fraudulent messages are recognized by analyzing the message length, out-of-range memory access attacks are therefore prevented.

## Examples
- **[Echo](https://github.com/qbeon/webwire-go/tree/master/examples/echo)** - Demonstrates a simple request-reply implementation.

- **[Pub-Sub](https://github.com/qbeon/webwire-go/tree/master/examples/pubsub)** - Demonstrantes a simple publisher-subscriber tolopology implementation.

- **[Chat Room](https://github.com/qbeon/webwire-go/tree/master/examples/chatroom)** - Demonstrates advanced use of the library. The corresponding [JavaScript Chat Room Client](https://github.com/qbeon/webwire-js/tree/master/examples/chatroom-client-vue) is implemented with the [Vue.js framework](https://vuejs.org/).

## Features
### Request-Reply
Clients can initiate multiple simultaneous requests and receive replies asynchronously. Requests are multiplexed through the connection similar to HTTP2 pipelining.

```go
// Send a request to the server, will block the goroutine until replied
reply, err := client.Request("", wwr.NewPayload(
  wwr.EncodingBinary,
  []byte("sudo rm -rf /"),
))
if err != nil {
  // Oh oh, request failed for some reason!
}
reply // Here we go!
 ```

Timed requests will timeout and return an error if the server doesn't manage to reply within the specified time frame.

```go
// Send a request to the server, will block the goroutine for 200ms at max
reply, err := client.TimedRequest("", wwr.Payload(
  wwr.EncodingUtf8,
  []byte("hurry up!"),
  200*time.Millisecond,
))
if err != nil {
  // Probably timed out!
}
reply // Just in time!
```

### Client-side Signals
Individual clients can send signals to the server. Signals are one-way messages guaranteed to arrive, though they're not guaranteed to be processed like requests are. In cases such as when the server is being shut down, incoming signals are ignored by the server and dropped while requests will acknowledge the failure.

```go
// Send signal to server
err := client.Signal(
  "eventA",
  wwr.NewPayload(
    wwr.EncodingUtf8,
    []byte("something"),
  ),
)
```

### Server-side Signals
The server also can send signals to individual connected clients.

```go
func OnRequest(
  _ context.Context,
  conn wwr.Connection,
  _ wwr.Message,
) (wwr.Payload, error) {
  // Send a signal to the client before replying to the request
  conn.Signal("", wwr.NewPayload(wwr.EncodingUtf8, []byte("ping!")))

  // Reply nothing
  return nil, nil
}
```

### Namespaces
Different kinds of requests and signals can be differentiated using the builtin namespacing feature.

```go
func OnRequest(
  _ context.Context,
  _ wwr.Connection,
  msg wwr.Message,
) (wwr.Payload, error) {
  switch msg.Name() {
  case "auth":
    // Authentication request
    return wwr.NewPayload(
      wwr.EncodingUtf8,
      []byte("this is an auth request"),
    )
  case "query":
    // Query request
    return wwr.NewPayload(
      wwr.EncodingUtf8,
      []byte("this is a query request"),
    )
  }

  // Otherwise return nothing
  return nil, nil
}
```
```go
func OnSignal(
  _ context.Context,
  _ wwr.Connection,
  msg wwr.Message,
) {
  switch msg.Name() {
  case "event A":
    // handle event A
  case "event B":
    // handle event B
  }
}
```

### Sessions
Individual connections can get sessions assigned to identify them. The state of the session is automagically synchronized between the client and the server. WebWire doesn't enforce any kind of authentication technique though, it just provides a way to authenticate a connection. WebWire also doesn't enforce any kind of session storage, the user could implement a custom session manager implementing the WebWire `SessionManager` interface to use any kind of volatile or persistent session storage, be it a database or a simple in-memory map.

```go
func OnRequest(
  _ context.Context,
  conn wwr.Connection,
  msg wwr.Message,
) (wwr.Payload, error) {
  // Verify credentials
  if string(msg.Payload().Data()) != "secret:pass" {
    return nil, wwr.ReqErr {
      Code: "WRONG_CREDENTIALS",
      Message: "Incorrect username or password, try again"
    }
  }
  // Create session (will automatically synchronize to the client)
  err := conn.CreateSession(/*something that implements wwr.SessionInfo*/)
  if err != nil {
    return nil, fmt.Errorf("Couldn't create session for some reason")
  }

  // Complete request, reply nothing
  return nil, nil
}
```

WebWire provides a basic file-based session manager implementation out of the box used by default when no custom session manager is defined. The default session manager creates a file with a .wwrsess extension for each opened session in the configured directory (which, by default, is the directory of the executable). During the restoration of a session the file is looked up by name using the session key, read and unmarshalled recreating the session object.

### Automatic Session Restoration
The client will automatically try to restore the previously opened session during connection establishment when getting disconnected without explicitly closing the session before.

```go
// Will automatically restore the previous session if there was any
err := client.Connect()
```

The session can also be restored manually given its key assuming the server didn't yet delete it. Session restoration will fail and return an error if the provided key doesn't correspond to any active session on the server or else if there's another active session already assigned to this client.
```go
err := client.RestoreSession([]byte("yoursessionkeygoeshere"))
```

### Automatic Connection Maintenance
The WebWire client maintains the connection fully automatically to guarantee maximum connection uptime. It will automatically reconnect in the background whenever the connection is lost.

The only things to remember are:
- Client API methods such as `client.Request`, `client.TimedRequest` and `client.RestoreSession` will timeout if the server is unavailable for the entire duration of the specified timeout and thus the client fails to reconnect.
- `client.Signal` will immediately return a `DisconnectedErr` error if there's no connection at the time the signal was sent.

This feature is entirely optional and can be disabled at will which will cause `client.Request`, `client.TimedRequest` and `client.RestoreSession` to immediately return a `DisconnectedErr` error when there's no connection at the time the request is made.

The WebWire server will also try to keep connections alive by periodically sending heartbeats to the client. The heartbeat interval and timeout durations are adjustable through the server options and default to 30 and 60 seconds respectively.

### Concurrency
Messages are parsed and handled concurrently in a separate goroutine by default. The total number of concurrently executed handlers can be throttled down to a specified number using the `MaxConcurrentHandlers` server option, which disables the throttling when set to `0`.

### Thread Safety
All exported interfaces provided by both the server and the client can safely be used concurrently from multiple goroutines, the library automatically synchronizes all concurrent operations.

### Hooks
Various hooks provide the ability to asynchronously react to different kinds of events and control the behavior of both the client and the server.

#### Server-side Hooks
- OnOptions
- BeforeUpgrade
- OnClientConnected
- OnClientDisconnected
- OnSignal
- OnRequest
- OnSessionKeyGeneration
- OnSessionCreated
- OnSessionLookup
- OnSessionClosed

#### Client-side Hooks
- OnServerSignal
- OnSessionCreated
- OnSessionClosed
- OnDisconnected

### Graceful Shutdown
The server will finish processing all ongoing signals and requests before closing when asked to shut down.
```go
// Will block the calling goroutine until all handlers have finished
server.Shutdown()
```
While the server is shutting down new connections are refused with `503 Service Unavailable` and incoming new requests from connected clients will be rejected with a special error: `RegErrSrvShutdown`. Any incoming signals from connected clients will be ignored during the shutdown.


### Seamless JavaScript Support
The [official JavaScript library](https://github.com/qbeon/webwire-js) enables seamless support for various JavaScript environments providing a fully compliant client implementation supporting the latest feature set of the [webwire-go](https://github.com/qbeon/webwire-go) library.

## Dependencies
This library depends on:
- **[websocket](https://github.com/gorilla/websocket)** version [v1.2.0](https://github.com/gorilla/websocket/releases/tag/v1.2.0) by **[Gorilla web toolkit](https://github.com/gorilla)** - A WebSocket implementation for Go.  
This library is used internally to abstract away the underlying websockets implementation.
- **[tmdwg-go](https://github.com/qbeon/tmdwg-go)** version [v1.0.0](https://github.com/qbeon/tmdwg-go/releases/tag/1.0.0) by **[QBEON](https://github.com/qbeon)** - A timed wait group implementation used internally for asynchronous testing.


----

© 2018 Roman Sharkov <roman.sharkov@qbeon.com>
