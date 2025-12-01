// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cbindings

/*
#include <stdlib.h>
#include "defra_structs.h"
*/
import "C"

import (
	"context"
	"time"
)

//export P2PInfo
func P2PInfo(nodePtr C.uintptr_t) C.Result {
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	addresses, err := node.DB.PeerInfo()
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(marshalJSONToGoCResult(addresses))
}

//export P2PActivePeers
func P2PActivePeers(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	peers, err := node.DB.ActivePeers(ctx)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(peers))
}

//export P2PgetAllReplicators
func P2PgetAllReplicators(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	reps, err := node.DB.GetAllReplicators(ctx)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(marshalJSONToGoCResult(reps))
}

//export P2PsetReplicator
func P2PsetReplicator(nodePtr C.uintptr_t, collections *C.char, addresses *C.char, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	colArgs := splitCommaSeparatedString(C.GoString(collections))
	addressesArgs := splitCommaSeparatedString(C.GoString(addresses))

	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	err = node.DB.SetReplicator(ctx, addressesArgs, colArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PdeleteReplicator
func P2PdeleteReplicator(nodePtr C.uintptr_t, collections *C.char, id *C.char, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	colArgs := splitCommaSeparatedString(C.GoString(collections))

	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	err = node.DB.DeleteReplicator(ctx, C.GoString(id), colArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PcollectionAdd
func P2PcollectionAdd(nodePtr C.uintptr_t, collections *C.char, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	colArgs := splitCommaSeparatedString(C.GoString(collections))

	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	err = node.DB.AddP2PCollections(ctx, colArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PcollectionRemove
func P2PcollectionRemove(nodePtr C.uintptr_t, collections *C.char, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	colArgs := splitCommaSeparatedString(C.GoString(collections))

	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	err = node.DB.RemoveP2PCollections(ctx, colArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PcollectionGetAll
func P2PcollectionGetAll(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	cols, err := node.DB.GetAllP2PCollections(ctx)

	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(marshalJSONToGoCResult(cols))
}

//export P2PdocumentAdd
func P2PdocumentAdd(nodePtr C.uintptr_t, collections *C.char, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	colArgs := splitCommaSeparatedString(C.GoString(collections))

	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	err = node.DB.AddP2PDocuments(ctx, colArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PdocumentRemove
func P2PdocumentRemove(nodePtr C.uintptr_t, collections *C.char, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	colArgs := splitCommaSeparatedString(C.GoString(collections))

	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	err = node.DB.RemoveP2PDocuments(ctx, colArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PdocumentGetAll
func P2PdocumentGetAll(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	cols, err := node.DB.GetAllP2PDocuments(ctx)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(marshalJSONToGoCResult(cols))
}

//export P2PdocumentSync
func P2PdocumentSync(nodePtr C.uintptr_t,
	collection *C.char,
	docIDs *C.char,
	timeoutStr *C.char,
	identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	docArgs := splitCommaSeparatedString(C.GoString(docIDs))
	timeoutDuration := time.Duration(0)

	timeout := C.GoString(timeoutStr)
	if timeout != "" {
		timeoutDurationParsed, err := time.ParseDuration(timeout)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		timeoutDuration = timeoutDurationParsed
	}

	if timeoutDuration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeoutDuration)
		defer cancel()
	}

	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	err = node.DB.SyncDocuments(ctx, C.GoString(collection), docArgs)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PcollectionSyncVersions
func P2PcollectionSyncVersions(nodePtr C.uintptr_t,
	versionIDs *C.char,
	timeoutStr *C.char,
	identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	versionArgs := splitCommaSeparatedString(C.GoString(versionIDs))
	timeoutDuration := time.Duration(0)

	timeout := C.GoString(timeoutStr)
	if timeout != "" {
		timeoutDurationParsed, err := time.ParseDuration(timeout)
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		timeoutDuration = timeoutDurationParsed
	}

	if timeoutDuration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeoutDuration)
		defer cancel()
	}

	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	err = node.DB.SyncCollectionVersions(ctx, versionArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PbranchableCollectionSync
func P2PbranchableCollectionSync(nodePtr C.uintptr_t,
	collectionID *C.char,
	timeoutStr *C.char,
	identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	if timeoutStr != nil {
		timeout, err := time.ParseDuration(C.GoString(timeoutStr))
		if err != nil {
			return returnC(returnGoC(1, err.Error(), ""))
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	err = node.DB.SyncBranchableCollection(ctx, C.GoString(collectionID))
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2Pconnect
func P2Pconnect(nodePtr C.uintptr_t, peerAddresses *C.char, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	addresses := splitCommaSeparatedString(C.GoString(peerAddresses))
	err = node.DB.Connect(ctx, addresses)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
