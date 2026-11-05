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

//go:build !js && !rust_ffi

package action

import (
	"errors"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/state"
)

// setupRustFFIClient reports that the Rust FFI client is not built in.
func setupRustFFIClient(
	_ *state.State,
	_ *node.Node,
	_ immutable.Option[state.Identity],
	_ NodeSetupConfig,
	_ string,
) (clients.Client, error) {
	return nil, errors.New("the Rust FFI client needs the rust_ffi build tag")
}
