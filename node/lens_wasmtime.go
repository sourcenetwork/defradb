// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js

package node

import (
	"github.com/lens-vm/lens/host-go/engine/module"
	"github.com/lens-vm/lens/host-go/runtimes/wasmtime"
)

const WasmTime LensRuntimeType = "wasm-time"

func init() {
	runtimeConstructors[DefaultLens] = func() module.Runtime { return wasmtime.New() }
	runtimeConstructors[WasmTime] = func() module.Runtime { return wasmtime.New() }
}
