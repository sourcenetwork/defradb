// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package http

import (
	"context"
	"net/http/httptest"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/node"
)

var _ client.TxnStore = (*Wrapper)(nil)
var _ client.P2P = (*Wrapper)(nil)

// Wrapper combines an HTTP client and server into a
// single struct that implements the client.TxnStore interface.
//
// *http.Client is embedded so its methods satisfy client.TxnStore and
// client.P2P without hand-written forwarding; only methods needing
// server-side state (NewTxn, Close, Events, ...) are defined explicitly.
type Wrapper struct {
	*http.Client
	node         *node.Node
	handler      *http.Handler
	httpServer   *httptest.Server
	serverCancel context.CancelFunc
}

func NewWrapper(node *node.Node) (*Wrapper, error) {
	handler, err := http.NewHandler(node.DB, node.Options())
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	handlerWithCtx := http.InjectServerContext(ctx)(handler)
	httpServer := httptest.NewServer(handlerWithCtx)
	client, err := http.NewClient(httpServer.URL)
	if err != nil {
		cancel()
		return nil, err
	}

	return &Wrapper{
		Client:       client,
		node:         node,
		handler:      handler,
		httpServer:   httpServer,
		serverCancel: cancel,
	}, nil
}

func (w *Wrapper) NewTxn(readOnly bool) (client.Txn, error) {
	clientTxn, err := w.Client.NewTxn(readOnly)
	if err != nil {
		return nil, err
	}
	serverTxn, err := w.handler.Transaction(clientTxn.ID())
	if err != nil {
		clientTxn.Discard()
		return nil, err
	}
	return &Transaction{Wrapper: w, clientTxn: clientTxn, txn: serverTxn}, nil
}

func (w *Wrapper) Close() {
	w.serverCancel()
	w.httpServer.Close()
	w.handler.Close()
	_ = w.node.Close(context.Background())
}

func (w *Wrapper) Events() event.Bus {
	return w.node.DB.Events()
}

func (w *Wrapper) MaxTxnRetries() int {
	return w.node.DB.MaxTxnRetries()
}

func (w *Wrapper) PrintDump(ctx context.Context) error {
	return w.node.DB.PrintDump(ctx)
}

func (w *Wrapper) Host() string {
	return w.httpServer.URL
}
