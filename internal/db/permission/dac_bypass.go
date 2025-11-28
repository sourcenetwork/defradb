// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package permission

import (
	"context"
	"sync"

	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
)

var identityBypassState = struct {
	cache map[string]bool
	mutex sync.Mutex
}{
	cache: make(map[string]bool),
}

// ClearIdentityFromBypassCache clears the identity from the cache if it exists,
// if it doesn't exist it is a no-op.
//
// Note: While this works right now for local node access control, if we were to
// ever implement global node access control this would not work. In that case we
// should probably either have a global cache of sorts for all nodes, or a simpler
// solution might be to have the bypass state computed per request (without cache).
func ClearIdentityFromBypassCache(identity string) {
	identityBypassState.mutex.Lock()
	defer identityBypassState.mutex.Unlock()
	delete(identityBypassState.cache, identity)
}

func canDACBypass(
	ctx context.Context,
	nodeACP NACInfo,
	identity string,
) bool {
	// Generally when NAC is not enabled, we allow the gated operations to work by assuming all
	// NAC permissions are granted, however allowing DAC bypass to work for everyone when NAC
	// is not enabled will defeat the purpose of having DAC, so don't bypass DAC in that case.
	if nodeACP.NodeACPDesc.Status != client.NACEnabled ||
		nodeACP.NodeACP == nil {
		return false
	}

	identityBypassState.mutex.Lock()
	defer identityBypassState.mutex.Unlock()

	hasDACBypass, exists := identityBypassState.cache[identity]
	if !exists { // unknown, so check access to bypass.
		err := CheckNodeOperationAccess(
			ctx,
			identity,
			nodeACP,
			acpTypes.NodeDACBypassPerm,
			acpTypes.NodeACPObject,
		)
		hasDACBypass = err == nil
		identityBypassState.cache[identity] = hasDACBypass
	}
	return hasDACBypass
}
