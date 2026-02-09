// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package state

import (
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
)

// registerIdentityIfNeeded is a no-op on js/wasm since the Rust FFI client
// is not available in that environment.
func registerIdentityIfNeeded(_ *State, _ acpIdentity.Identity) {}
