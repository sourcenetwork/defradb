// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"github.com/lens-vm/lens/host-go/engine/module"
	"github.com/lens-vm/lens/host-go/runtimes/js"
)

const JSLensRuntime LensRuntimeType = "js"

func init() {
	runtimeConstructors[DefaultLens] = func() module.Runtime { return js.New() }
	runtimeConstructors[JSLensRuntime] = func() module.Runtime { return js.New() }
}
