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

	"github.com/sourcenetwork/defradb/internal/db"
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
// acpType is optional and can be:
// - "sourcehub" to use SourceHub ACP
// - anything else (including undefined/null) to use Local ACP
func open(this js.Value, args []js.Value) (js.Value, error) {
	var acpType string
	if len(args) > 0 && args[0].Type() == js.TypeString {
		acpType = args[0].String()
	}
	ident, err := initKeypairAndGetIdentity()
	if err != nil {
		return js.Undefined(), err
	}
	opts := []node.Option{
		node.WithStoreType(node.BadgerStore),
		node.WithBadgerInMemory(true),
		node.WithDisableP2P(true),
		node.WithDisableAPI(true),
		db.WithNodeIdentity(ident),
	}
	if acpType == "sourcehub" {
		opts = append(opts, node.WithDocumentACPType(node.SourceHubJsDocumentACPType))
	}
	n, err := node.New(context.Background(), opts...)
	if err != nil {
		return js.Undefined(), err
	}
	if err := n.Start(context.Background()); err != nil {
		return js.Undefined(), err
	}
	return NewClient(n).JSValue(), nil
}
