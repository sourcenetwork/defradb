// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build cgo
// +build cgo

package cwrap

/*
#include "defra_structs.h"
*/
import "C"

import (
	"context"
	"encoding/json"

	"github.com/libp2p/go-libp2p/core/peer"
)

func P2PInfo() *C.Result {
	info := globalNode.Peer.PeerInfo()
	return marshalJSONToCResult(info)
}

func P2PgetAllReplicators() *C.Result {
	ctx := context.Background()
	reps, err := globalNode.Peer.GetAllReplicators(ctx)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return marshalJSONToCResult(reps)
}

func P2PsetReplicator(cCollections *C.char, cPeer *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	peerStr := C.GoString(cPeer)
	colArgs := splitCommaSeparatedCString(cCollections)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	var info peer.AddrInfo
	if err := json.Unmarshal([]byte(peerStr), &info); err != nil {
		return returnC(1, err.Error(), "")
	}

	// Set the replicator and return the result
	err = globalNode.Peer.SetReplicator(ctx, info, colArgs...)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}

func P2PdeleteReplicator(cCollections *C.char, cPeer *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	peerStr := C.GoString(cPeer)
	colArgs := splitCommaSeparatedCString(cCollections)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	var info peer.AddrInfo
	if err := json.Unmarshal([]byte(peerStr), &info); err != nil {
		return returnC(1, err.Error(), "")
	}

	// Set the replicator and return the result
	err = globalNode.Peer.DeleteReplicator(ctx, info, colArgs...)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}

func P2PcollectionAdd(cCollections *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	colArgs := splitCommaSeparatedCString(cCollections)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Try to add the collections, then return the result
	err = globalNode.Peer.AddP2PCollections(ctx, colArgs...)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}

func P2PcollectionRemove(cCollections *C.char, cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()
	colArgs := splitCommaSeparatedCString(cCollections)

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Try to remove the collections, then return the result
	err = globalNode.Peer.RemoveP2PCollections(ctx, colArgs...)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return returnC(0, "", "")
}

func P2PcollectionGetAll(cTxnID C.ulonglong) *C.Result {
	ctx := context.Background()

	// Set the transaction
	newctx, err := contextWithTransaction(ctx, cTxnID)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	ctx = newctx

	// Try to get the collections, then return
	cols, err := globalNode.Peer.GetAllP2PCollections(ctx)
	if err != nil {
		return returnC(1, err.Error(), "")
	}
	return marshalJSONToCResult(cols)
}
