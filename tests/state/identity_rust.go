// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js && rust_ffi

package state

import (
	"github.com/stretchr/testify/require"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/tests/clients/rustffi"
)

// registerIdentityIfNeeded registers the identity with the Rust FFI layer
// when the Rust FFI client is in use. This ensures block signing works correctly.
func registerIdentityIfNeeded(s *State, identity acpIdentity.Identity) {
	if s.ClientType == RustFFIClientType {
		err := rustffi.RegisterIdentityWithRust(identity)
		require.NoError(s.T, err, "failed to register identity with Rust FFI")
	}
}
