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

//export GetP2PInfo
func GetP2PInfo(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	opts := options.WithIdentity(options.PeerInfo(), acpIdentity.FromContext(ctx))
	addresses, err := node.DB.PeerInfo(ctx, opts)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(marshalJSONToGoCResult(addresses))
}

//export ListP2PActivePeers
func ListP2PActivePeers(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	opts := options.WithIdentity(options.ActivePeers(), acpIdentity.FromContext(ctx))
	peers, err := node.DB.ActivePeers(ctx, opts)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	return returnC(marshalJSONToGoCResult(peers))
}

//export ListP2PReplicators
func ListP2PReplicators(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	listRepOpt := options.WithIdentity(options.ListReplicators(), acpIdentity.FromContext(ctx))
	reps, err := node.DB.ListReplicators(ctx, listRepOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(marshalJSONToGoCResult(reps))
}

//export AddP2PReplicator
func AddP2PReplicator(nodePtr C.uintptr_t,
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

	ctx = attachTxnFromPointer(nodePtr, ctx)

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

//export DeleteP2PReplicator
func DeleteP2PReplicator(nodePtr C.uintptr_t, collections *C.char, id *C.char, identityPtr C.uintptr_t) C.Result {
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

	ctx = attachTxnFromPointer(nodePtr, ctx)

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

//export AddP2PCollection
func AddP2PCollection(nodePtr C.uintptr_t, collections *C.char, identityPtr C.uintptr_t) C.Result {
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

	ctx = attachTxnFromPointer(nodePtr, ctx)

	addP2PColOpt := options.WithIdentity(options.AddP2PCollections(), acpIdentity.FromContext(ctx))
	err = node.DB.AddP2PCollections(ctx, colArgs, addP2PColOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export DeleteP2PCollection
func DeleteP2PCollection(nodePtr C.uintptr_t, collections *C.char, identityPtr C.uintptr_t) C.Result {
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

	ctx = attachTxnFromPointer(nodePtr, ctx)

	deleteP2PColOpt := options.WithIdentity(options.DeleteP2PCollections(), acpIdentity.FromContext(ctx))
	err = node.DB.DeleteP2PCollections(ctx, colArgs, deleteP2PColOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export ListP2PCollections
func ListP2PCollections(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	listP2PColOpt := options.WithIdentity(options.ListP2PCollections(), acpIdentity.FromContext(ctx))
	cols, err := node.DB.ListP2PCollections(ctx, listP2PColOpt)

	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(marshalJSONToGoCResult(cols))
}

//export AddP2PDocument
func AddP2PDocument(nodePtr C.uintptr_t, collections *C.char, identityPtr C.uintptr_t) C.Result {
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

	ctx = attachTxnFromPointer(nodePtr, ctx)

	addP2PDocOpt := options.WithIdentity(options.AddP2PDocuments(), acpIdentity.FromContext(ctx))
	err = node.DB.AddP2PDocuments(ctx, colArgs, addP2PDocOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export DeleteP2PDocument
func DeleteP2PDocument(nodePtr C.uintptr_t, collections *C.char, identityPtr C.uintptr_t) C.Result {
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

	ctx = attachTxnFromPointer(nodePtr, ctx)

	deleteP2PDocOpt := options.WithIdentity(options.DeleteP2PDocuments(), acpIdentity.FromContext(ctx))
	err = node.DB.DeleteP2PDocuments(ctx, colArgs, deleteP2PDocOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export ListP2PDocuments
func ListP2PDocuments(nodePtr C.uintptr_t, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	listP2PDocOpt := options.WithIdentity(options.ListP2PDocuments(), acpIdentity.FromContext(ctx))
	cols, err := node.DB.ListP2PDocuments(ctx, listP2PDocOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(marshalJSONToGoCResult(cols))
}

//export SyncP2PDocuments
func SyncP2PDocuments(nodePtr C.uintptr_t,
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

	ctx = attachTxnFromPointer(nodePtr, ctx)

	syncOpts := options.WithIdentity(options.SyncDocuments(), acpIdentity.FromContext(ctx))
	err = node.DB.SyncDocuments(ctx, C.GoString(collection), docArgs, syncOpts)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export SyncP2PCollectionVersions
func SyncP2PCollectionVersions(nodePtr C.uintptr_t,
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

	ctx = attachTxnFromPointer(nodePtr, ctx)

	opts := options.WithIdentity(options.SyncCollectionVersions(), acpIdentity.FromContext(ctx))
	err = node.DB.SyncCollectionVersions(ctx, versionArgs, opts)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export SyncP2PBranchableCollection
func SyncP2PBranchableCollection(nodePtr C.uintptr_t,
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

	ctx = attachTxnFromPointer(nodePtr, ctx)

	opts := options.WithIdentity(options.SyncBranchableCollection(), acpIdentity.FromContext(ctx))
	err = node.DB.SyncBranchableCollection(ctx, C.GoString(collectionID), opts)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export ConnectP2PPeers
func ConnectP2PPeers(nodePtr C.uintptr_t, peerAddresses *C.char, identityPtr C.uintptr_t) C.Result {
	ctx := context.Background()
	ctx, err := contextWithIdentity(ctx, identityPtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	ctx = attachTxnFromPointer(nodePtr, ctx)

	addresses := splitCommaSeparatedString(C.GoString(peerAddresses))
	connectOpt := options.WithIdentity(options.Connect(), acpIdentity.FromContext(ctx))
	err = node.DB.Connect(ctx, addresses, connectOpt)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
