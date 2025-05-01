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

	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/goji"
)

// SetGlobal sets the global defradb variable so that it is
// accessible from any JS context (browser, wasm, etc.)
func SetGlobal() {
	js.Global().Set("defradb", map[string]any{
		"open": goji.Async(open),
	})
}

// open creates a new DB client and returns it wrapped in a JS object.
func open(this js.Value, args []js.Value) (js.Value, error) {
	opts := []node.Option{
		node.WithStoreType(node.BadgerStore),
		node.WithBadgerInMemory(true),
		node.WithDisableP2P(true),
		node.WithDisableAPI(true),
	}
	n, err := node.New(context.Background(), opts...)
	if err != nil {
		return js.Undefined(), err
	}
	err = n.Start(context.Background())
	if err != nil {
		return js.Undefined(), err
	}
	return NewClient(n).JSValue(), nil
}
