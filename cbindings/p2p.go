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

	"github.com/sourcenetwork/defradb/client/options"
	acpIdentity "github.com/sourcenetwork/defradb/internal/identity"
)

//export P2PInfo
func P2PInfo(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	opts := options.WithIdentity(options.PeerInfo(), acpIdentity.FromContext(ctx))
	addresses, err := node.DB.PeerInfo(ctx, opts)
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

	opts := options.WithIdentity(options.ActivePeers(), acpIdentity.FromContext(ctx))
	peers, err := node.DB.ActivePeers(ctx, opts)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(peers))
}

//export P2PreplicatorList
func P2PreplicatorList(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	listRepOpt := options.WithIdentity(options.ListReplicators(), acpIdentity.FromContext(ctx))
	reps, err := node.DB.ListReplicators(ctx, listRepOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(marshalJSONToGoCResult(reps))
}

//export P2PreplicatorAdd
func P2PreplicatorAdd(nodePtr C.uintptr_t,
	collections *C.char,
	addresses *C.char,
	identityPtr C.uintptr_t) C.Result {
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
	opt := options.WithIdentity(
		options.AddReplicator().SetCollectionNames(colArgs),
		acpIdentity.FromContext(ctx),
	)
	err = node.DB.AddReplicator(ctx, addressesArgs, opt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PreplicatorDelete
func P2PreplicatorDelete(nodePtr C.uintptr_t, collections *C.char, id *C.char, identityPtr C.uintptr_t) C.Result {
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
	delRepOpt := options.WithIdentity(
		options.DeleteReplicator().SetCollectionNames(colArgs),
		acpIdentity.FromContext(ctx),
	)
	err = node.DB.DeleteReplicator(ctx, C.GoString(id), delRepOpt)
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
	createP2PColOpt := options.WithIdentity(options.AddP2PCollections(), acpIdentity.FromContext(ctx))
	err = node.DB.AddP2PCollections(ctx, colArgs, createP2PColOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PcollectionDelete
func P2PcollectionDelete(nodePtr C.uintptr_t, collections *C.char, identityPtr C.uintptr_t) C.Result {
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
	deleteP2PColOpt := options.WithIdentity(options.DeleteP2PCollections(), acpIdentity.FromContext(ctx))
	err = node.DB.DeleteP2PCollections(ctx, colArgs, deleteP2PColOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PcollectionList
func P2PcollectionList(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	listP2PColOpt := options.WithIdentity(options.ListP2PCollections(), acpIdentity.FromContext(ctx))
	cols, err := node.DB.ListP2PCollections(ctx, listP2PColOpt)

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
	createP2PDocOpt := options.WithIdentity(options.AddP2PDocuments(), acpIdentity.FromContext(ctx))
	err = node.DB.AddP2PDocuments(ctx, colArgs, createP2PDocOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PdocumentDelete
func P2PdocumentDelete(nodePtr C.uintptr_t, collections *C.char, identityPtr C.uintptr_t) C.Result {
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
	deleteP2PDocOpt := options.WithIdentity(options.DeleteP2PDocuments(), acpIdentity.FromContext(ctx))
	err = node.DB.DeleteP2PDocuments(ctx, colArgs, deleteP2PDocOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PdocumentList
func P2PdocumentList(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	listP2PDocOpt := options.WithIdentity(options.ListP2PDocuments(), acpIdentity.FromContext(ctx))
	cols, err := node.DB.ListP2PDocuments(ctx, listP2PDocOpt)
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
	opts := options.WithIdentity(options.SyncCollectionVersions(), acpIdentity.FromContext(ctx))
	err = node.DB.SyncCollectionVersions(ctx, versionArgs, opts)
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
	opts := options.WithIdentity(options.SyncBranchableCollection(), acpIdentity.FromContext(ctx))
	err = node.DB.SyncBranchableCollection(ctx, C.GoString(collectionID), opts)
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
	connectOpt := options.WithIdentity(options.Connect(), acpIdentity.FromContext(ctx))
	err = node.DB.Connect(ctx, addresses, connectOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
