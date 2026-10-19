// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build js

package js

import (
	"context"
	"syscall/js"

	"github.com/sourcenetwork/goji"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/node"
)

// SetGlobal sets the global defradb variable so that it is
// accessible from any JS context (browser, wasm, etc.)
func SetGlobal() {
	js.Global().Set("defradb", map[string]any{
		"open": goji.Async(open),
	})
}

// open creates a new DB client and returns it wrapped in a JS object.
//
// The first argument is an optional object containing a field for each node.Option.
func open(this js.Value, args []js.Value) (js.Value, error) {
	optsVal := optionsValue(args, 0)

	nodeOpts := node.DefaultNodeOptions()
	nodeOpts.Store.Path = "/defradb"
	nodeOpts.Store.Store = options.NodeStoreType("level")
	nodeOpts.P2P.ListenAddresses = []string{"/ip4/0.0.0.0/udp/0/quic-v1/webtransport"}

	if err := parseNodeOptions(optsVal, &nodeOpts); err != nil {
		return js.Undefined(), err
	}
	n, err := node.New(context.Background(), asOpts(nodeOpts))
	if err != nil {
		return js.Undefined(), err
	}
	if err := n.Start(context.Background()); err != nil {
		_ = n.Close(context.Background())
		return js.Undefined(), err
	}
	return NewClient(n).JSValue(), nil
}
