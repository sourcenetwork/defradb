// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js && !rust_ffi

package tests

import (
	"fmt"

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/state"
)

// setupRustFFIClient is a stub that returns an error when the rust_ffi build tag is not set.
func setupRustFFIClient(
	_ *state.State,
	_ *node.Node,
	_ immutable.Option[acpIdentity.Identity],
	_ int,
	_ bool,
	_ string,
) (clients.Client, error) {
	return nil, fmt.Errorf("rust FFI client not available: build with -tags rust_ffi")
}
