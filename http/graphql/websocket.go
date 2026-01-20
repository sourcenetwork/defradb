// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// primarily based on https://github.com/99designs/gqlgen/blob/master/graphql/handler/transport/websocket.go
package graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/sourcenetwork/defradb/client"
)

type (
	Websocket struct {
		AllowedOrigins        []string
		Upgrader              websocket.Upgrader
		InitFunc              WebsocketInitFunc
		InitTimeout           time.Duration
		ErrorFunc             WebsocketErrorFunc
		CloseFunc             WebsocketCloseFunc
		KeepAlivePingInterval time.Duration
		PongOnlyInterval      time.Duration
		PingPongInterval      time.Duration
		/* If PingPongInterval has a non-0 duration, then when the server sends a ping
		 * it sets a ReadDeadline of PingPongInterval*2 and if the client doesn't respond
		 * with pong before that deadline is reached then the connection will die with a
		 * 1006 error code.
		 *
		 * MissingPongOk if true, tells the server to not use a ReadDeadline such that a
		 * missing/slow pong response from the client doesn't kill the connection.
		 */
		MissingPongOk bool

		didInjectSubprotocols bool
	}
	wsConnection struct {
		Websocket
		ctx             context.Context
		conn            *websocket.Conn
		me              messageExchanger
		active          map[string]context.CancelFunc
		mu              sync.Mutex
		keepAliveTicker *time.Ticker
		pongOnlyTicker  *time.Ticker
		pingPongTicker  *time.Ticker
		receivedPong    bool
		exec            Executor
		closed          bool
		headers         http.Header

		initPayload InitPayload
	}

	WebsocketInitFunc  func(ctx context.Context, initPayload InitPayload) (context.Context, *InitPayload, error)
	WebsocketErrorFunc func(ctx context.Context, err error)

	// Callback called when websocket is closed.
	WebsocketCloseFunc func(ctx context.Context, closeCode int)
)

type WebsocketError struct {
	Err error

	// IsReadError flags whether the error occurred on read or write to the websocket
	IsReadError bool
}

func (e WebsocketError) Error() string {
	if e.IsReadError {
		return fmt.Sprintf("websocket read: %v", e.Err)
	}
	return fmt.Sprintf("websocket write: %v", e.Err)
}

var (
	_ Transport = Websocket{}
	_ error     = WebsocketError{}
)

func (t Websocket) Supports(r *http.Request) bool {
	return r.Header.Get("Upgrade") != ""
}

func (t Websocket) Methods() []string {
	return []string{http.MethodGet}
}

func (t Websocket) Do(w http.ResponseWriter, r *http.Request, exec Executor) {
	t.injectGraphQLWSAllowedOrigins(t.AllowedOrigins)
	t.injectGraphQLWSSubprotocols()
	ws, err := t.Upgrader.Upgrade(w, r, http.Header{})
	if err != nil {
		log.ErrorE("unable to upgrade to websocket", err)
		responseJSON(w, http.StatusBadRequest, errorResponse{ErrUnableToUpgrade})
		return
	}

	var me messageExchanger
	switch ws.Subprotocol() {
	default:
		msg := websocket.FormatCloseMessage(
			websocket.CloseProtocolError,
			fmt.Sprintf("unsupported negotiated subprotocol %s", ws.Subprotocol()),
		)
		_ = ws.WriteMessage(websocket.CloseMessage, msg)
		return
	case graphqlwsSubprotocol, "":
		// clients are required to send a subprotocol, to be backward compatible with the previous
		// implementation we select
		// "graphql-ws" by default
		me = graphqlwsMessageExchanger{c: ws}
	case graphqltransportwsSubprotocol:
		me = graphqltransportwsMessageExchanger{c: ws}
	}

	conn := wsConnection{
		active:    map[string]context.CancelFunc{},
		conn:      ws,
		ctx:       r.Context(),
		exec:      exec,
		me:        me,
		headers:   r.Header,
		Websocket: t,
	}

	if !conn.init() {
		return
	}

	conn.run()
}

const (
	initMessageType messageType = iota
	connectionAckMessageType
	keepAliveMessageType
	connectionErrorMessageType
	connectionCloseMessageType
	startMessageType
	stopMessageType
	dataMessageType
	completeMessageType
	errorMessageType
	pingMessageType
	pongMessageType
)

var (
	supportedSubprotocols = []string{
		graphqlwsSubprotocol,
		graphqltransportwsSubprotocol,
	}
)

type (
	messageType int
	message     struct {
		payload json.RawMessage
		id      string
		t       messageType
	}
	messageExchanger interface {
		NextMessage() (message, error)
		Send(m *message) error
	}
)

func (t messageType) String() string {
	var text string
	switch t {
	default:
		text = "unknown"
	case initMessageType:
		text = "init"
	case connectionAckMessageType:
		text = "connection ack"
	case keepAliveMessageType:
		text = "keep alive"
	case connectionErrorMessageType:
		text = "connection error"
	case connectionCloseMessageType:
		text = "connection close"
	case startMessageType:
		text = "start"
	case stopMessageType:
		text = "stop subscription"
	case dataMessageType:
		text = "data"
	case completeMessageType:
		text = "complete"
	case errorMessageType:
		text = "error"
	case pingMessageType:
		text = "ping"
	case pongMessageType:
		text = "pong"
	}
	return text
}

func (t *Websocket) injectGraphQLWSAllowedOrigins(allowedOrigins []string) {
	// If no allowed origins are configured, leave CheckOrigin nil to preserve
	// gorilla/websocket's default same-origin policy.
	if len(allowedOrigins) == 0 {
		return
	}
	t.Upgrader.CheckOrigin = func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if slices.Contains(allowedOrigins, "*") {
			return true
		}
		return slices.Contains(allowedOrigins, strings.ToLower(origin))
	}
}

func (t *Websocket) injectGraphQLWSSubprotocols() {
	// the list of subprotocols is specified by the consumer of the Websocket struct,
	// in order to preserve backward compatibility, we inject the graphql specific subprotocols
	// at runtime
	if !t.didInjectSubprotocols {
		defer func() {
			t.didInjectSubprotocols = true
		}()

		for _, subprotocol := range supportedSubprotocols {
			if !contains(t.Upgrader.Subprotocols, subprotocol) {
				t.Upgrader.Subprotocols = append(t.Upgrader.Subprotocols, subprotocol)
			}
		}
	}
}

func (c *wsConnection) handlePossibleError(err error, isReadError bool) {
	if c.ErrorFunc != nil && err != nil {
		c.ErrorFunc(c.ctx, WebsocketError{
			Err:         err,
			IsReadError: isReadError,
		})
	}
}

func (c *wsConnection) nextMessageWithTimeout(timeout time.Duration) (message, error) {
	messages, errs := make(chan message, 1), make(chan error, 1)

	go func() {
		if m, err := c.me.NextMessage(); err != nil {
			errs <- err
		} else {
			messages <- m
		}
	}()

	select {
	case m := <-messages:
		return m, nil
	case err := <-errs:
		return message{}, err
	case <-time.After(timeout):
		return message{}, ErrReadTimeout
	}
}

func (c *wsConnection) init() bool {
	var m message
	var err error

	if c.InitTimeout != 0 {
		m, err = c.nextMessageWithTimeout(c.InitTimeout)
	} else {
		m, err = c.me.NextMessage()
	}

	if err != nil {
		if errors.Is(err, ErrReadTimeout) {
			c.close(websocket.CloseProtocolError, "connection initialisation timeout")
			return false
		}

		if errors.Is(err, ErrInvalidMsg) {
			c.sendConnectionError("invalid json")
		}

		c.close(websocket.CloseProtocolError, "decoding error")
		return false
	}

	switch m.t {
	case initMessageType:
		if len(m.payload) > 0 {
			c.initPayload = make(InitPayload)
			err := json.Unmarshal(m.payload, &c.initPayload)
			if err != nil {
				return false
			}
		}

		var initAckPayload *InitPayload
		if c.InitFunc != nil {
			var ctx context.Context
			ctx, initAckPayload, err = c.InitFunc(c.ctx, c.initPayload)
			if err != nil {
				c.sendConnectionError("%s", err.Error())
				c.close(websocket.CloseNormalClosure, "terminated")
				return false
			}
			c.ctx = ctx
		}

		if initAckPayload != nil {
			initJsonAckPayload, err := json.Marshal(*initAckPayload)
			if err != nil {
				log.ErrorE("failed to marshal init ack payload", err)
				return false
			}
			c.write(&message{t: connectionAckMessageType, payload: initJsonAckPayload})
		} else {
			c.write(&message{t: connectionAckMessageType})
		}
		c.write(&message{t: keepAliveMessageType})
	case connectionCloseMessageType:
		c.close(websocket.CloseNormalClosure, "terminated")
		return false
	default:
		c.sendConnectionError("unexpected message %s", m.t)
		c.close(websocket.CloseProtocolError, "unexpected message")
		return false
	}

	return true
}

func (c *wsConnection) write(msg *message) {
	c.mu.Lock()
	c.handlePossibleError(c.me.Send(msg), false)
	c.mu.Unlock()
}

func (c *wsConnection) run() {
	// We create a cancellation that will shutdown the keep-alive when we leave
	// this function.
	ctx, cancel := context.WithCancel(c.ctx)
	defer func() {
		cancel()
	}()

	// If we're running in graphql-ws mode, create a timer that will trigger a
	// keep alive message every interval
	if (c.conn.Subprotocol() == "" || c.conn.Subprotocol() == graphqlwsSubprotocol) &&
		c.KeepAlivePingInterval != 0 {
		c.mu.Lock()
		c.keepAliveTicker = time.NewTicker(c.KeepAlivePingInterval)
		c.mu.Unlock()

		go c.keepAlive(ctx)
	}

	// If we're running in graphql-transport-ws mode, create a timer that will trigger a
	// just a pong message every interval
	if c.conn.Subprotocol() == graphqltransportwsSubprotocol && c.PongOnlyInterval != 0 {
		c.mu.Lock()
		c.pongOnlyTicker = time.NewTicker(c.PongOnlyInterval)
		c.mu.Unlock()

		go c.keepAlivePongOnly(ctx)
	}

	// If we're running in graphql-transport-ws mode, create a timer that will
	// trigger a ping message every interval and expect a pong!
	if c.conn.Subprotocol() == graphqltransportwsSubprotocol && c.PingPongInterval != 0 {
		c.mu.Lock()
		c.pingPongTicker = time.NewTicker(c.PingPongInterval)
		c.mu.Unlock()

		if !c.MissingPongOk {
			// Note: when the connection is closed by this deadline, the client
			// will receive an "invalid close code"
			_ = c.conn.SetReadDeadline(time.Now().UTC().Add(2 * c.PingPongInterval))
		}
		go c.ping(ctx)
	}

	// Close the connection when the context is cancelled.
	// Will optionally send a "close reason" that is retrieved from the context.
	go c.closeOnCancel(ctx)

	for {
		m, err := c.me.NextMessage()
		if err != nil {
			// If the connection got closed by us, don't report the error
			if !errors.Is(err, net.ErrClosed) {
				c.handlePossibleError(err, true)
			}
			return
		}

		switch m.t {
		case startMessageType:
			c.subscribe(&m)
		case stopMessageType:
			c.mu.Lock()
			closer := c.active[m.id]
			c.mu.Unlock()
			if closer != nil {
				closer()
			}
		case connectionCloseMessageType:
			c.close(websocket.CloseNormalClosure, "terminated")
			return
		case pingMessageType:
			c.write(&message{t: pongMessageType, payload: m.payload})
		case pongMessageType:
			c.mu.Lock()
			c.receivedPong = true
			c.mu.Unlock()
			// Clear ReadTimeout -- 0 time val clears.
			_ = c.conn.SetReadDeadline(time.Time{})
		default:
			c.sendConnectionError("unexpected message %s", m.t)
			c.close(websocket.CloseProtocolError, "unexpected message")
			return
		}
	}
}

func (c *wsConnection) keepAlivePongOnly(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.pongOnlyTicker.Stop()
			return
		case <-c.pongOnlyTicker.C:
			c.write(&message{t: pongMessageType, payload: json.RawMessage{}})
		}
	}
}

func (c *wsConnection) keepAlive(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.keepAliveTicker.Stop()
			return
		case <-c.keepAliveTicker.C:
			c.write(&message{t: keepAliveMessageType})
		}
	}
}

func (c *wsConnection) ping(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.pingPongTicker.Stop()
			return
		case <-c.pingPongTicker.C:
			c.write(&message{t: pingMessageType, payload: json.RawMessage{}})
			// The initial deadline for this method is set in run()
			// if we have not yet received a pong, don't reset the deadline.
			c.mu.Lock()
			if !c.MissingPongOk && c.receivedPong {
				_ = c.conn.SetReadDeadline(time.Now().UTC().Add(2 * c.PingPongInterval))
			}
			c.receivedPong = false
			c.mu.Unlock()
		}
	}
}

func (c *wsConnection) closeOnCancel(ctx context.Context) {
	<-ctx.Done()

	if r := closeReasonForContext(ctx); r != "" {
		c.sendConnectionError("%s", r)
	}
	c.close(websocket.CloseNormalClosure, "terminated")
}

func (c *wsConnection) subscribe(msg *message) {
	var request Request
	if err := json.NewDecoder(bytes.NewReader(msg.payload)).Decode(&request); err != nil {
		c.sendError(msg.id, err)
		c.complete(msg.id)
		return
	}

	var options []client.RequestOption
	if request.OperationName != "" {
		options = append(options, client.WithOperationName(request.OperationName))
	}
	if len(request.Variables) > 0 {
		options = append(options, client.WithVariables(request.Variables))
	}

	ctx := c.ctx
	if c.initPayload != nil {
		ctx = withInitPayload(ctx, c.initPayload)
	}

	ctx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	c.active[msg.id] = cancel
	c.mu.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				var err error
				if e, ok := r.(error); ok {
					err = e
				} else {
					err = fmt.Errorf("%v", r)
				}
				c.sendError(msg.id, err)
			}
			c.complete(msg.id)
			c.mu.Lock()
			delete(c.active, msg.id)
			c.mu.Unlock()
			cancel()
		}()

		result := c.exec(ctx, request.Query, options...)

		// Handle immediate errors (non-subscription queries or errors)
		if len(result.GQL.Errors) > 0 {
			c.sendResponse(msg.id, result.GQL)
			return
		}

		// If this is not a subscription, send the result and complete
		if result.Subscription == nil {
			c.sendResponse(msg.id, result.GQL)
			return
		}

		// Handle subscription - iterate over the channel like SSE does
		for {
			select {
			case <-ctx.Done():
				return
			case item, open := <-result.Subscription:
				if !open {
					return
				}
				c.sendResponse(msg.id, item)
			}
		}
	}()
}

func (c *wsConnection) sendResponse(id string, response client.GQLResult) {
	b, err := json.Marshal(response)
	if err != nil {
		log.ErrorE("failed to marshal response", err)
		return
	}
	c.write(&message{
		payload: b,
		id:      id,
		t:       dataMessageType,
	})
}

func (c *wsConnection) complete(id string) {
	c.write(&message{id: id, t: completeMessageType})
}

func (c *wsConnection) sendError(id string, errs ...error) {
	// Per GraphQL spec, errors should be an array of error objects
	errList := make([]map[string]any, len(errs))
	for i, err := range errs {
		errList[i] = map[string]any{"message": err.Error()}
	}
	b, err := json.Marshal(errList)
	if err != nil {
		log.ErrorE("failed to marshal error", err)
		return
	}
	c.write(&message{t: errorMessageType, id: id, payload: b})
}

func (c *wsConnection) sendConnectionError(format string, args ...any) {
	errPayload := map[string]any{
		"message": fmt.Sprintf(format, args...),
	}
	b, err := json.Marshal(errPayload)
	if err != nil {
		log.ErrorE("failed to marshal connection error", err)
		return
	}

	c.write(&message{t: connectionErrorMessageType, payload: b})
}

func (c *wsConnection) close(closeCode int, message string) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	_ = c.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(closeCode, message),
	)
	for _, closer := range c.active {
		closer()
	}
	c.closed = true
	c.mu.Unlock()
	_ = c.conn.Close()

	if c.CloseFunc != nil {
		c.CloseFunc(c.ctx, closeCode)
	}
}

func contains(list []string, elem string) bool {
	for _, e := range list {
		if e == elem {
			return true
		}
	}

	return false
}

func handleNextReaderError(err error) error {
	// TODO: should we consider all closure scenarios here for the ws connection?
	// for now we only list the error codes from the previous implementation
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
		return ErrWsConnClosed
	}

	return err
}
