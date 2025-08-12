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
	"encoding/json"
	"runtime/cgo"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/defradb/node"
)

//export P2PInfo
func P2PInfo(nodePtr C.uintptr_t) *C.Result {
	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	info := node.Peer.PeerInfo()
	return returnC(marshalJSONToGoCResult(info))
}

//export P2PgetAllReplicators
func P2PgetAllReplicators(nodePtr C.uintptr_t) *C.Result {
	ctx := context.Background()
	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	reps, err := node.Peer.GetAllReplicators(ctx)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(marshalJSONToGoCResult(reps))
}

//export P2PsetReplicator
func P2PsetReplicator(nodePtr C.uintptr_t, collections *C.char, peerInfo *C.char) *C.Result {
	ctx := context.Background()
	colArgs := splitCommaSeparatedString(C.GoString(collections))

	var info peer.AddrInfo
	if err := json.Unmarshal([]byte(C.GoString(peerInfo)), &info); err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	err := node.Peer.SetReplicator(ctx, info, colArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PdeleteReplicator
func P2PdeleteReplicator(nodePtr C.uintptr_t, collections *C.char, peerInfo *C.char) *C.Result {
	ctx := context.Background()
	colArgs := splitCommaSeparatedString(C.GoString(collections))

	var info peer.AddrInfo
	if err := json.Unmarshal([]byte(C.GoString(peerInfo)), &info); err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	err := node.Peer.DeleteReplicator(ctx, info, colArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PcollectionAdd
func P2PcollectionAdd(nodePtr C.uintptr_t, collections *C.char) *C.Result {
	ctx := context.Background()
	colArgs := splitCommaSeparatedString(C.GoString(collections))

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	err := node.Peer.AddP2PCollections(ctx, colArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PcollectionRemove
func P2PcollectionRemove(nodePtr C.uintptr_t, collections *C.char) *C.Result {
	ctx := context.Background()
	colArgs := splitCommaSeparatedString(C.GoString(collections))

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	err := node.Peer.RemoveP2PCollections(ctx, colArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PcollectionGetAll
func P2PcollectionGetAll(nodePtr C.uintptr_t) *C.Result {
	ctx := context.Background()

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	cols, err := node.Peer.GetAllP2PCollections(ctx)

	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(marshalJSONToGoCResult(cols))
}

//export P2PdocumentAdd
func P2PdocumentAdd(nodePtr C.uintptr_t, collections *C.char) *C.Result {
	ctx := context.Background()
	colArgs := splitCommaSeparatedString(C.GoString(collections))

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	err := node.Peer.AddP2PDocuments(ctx, colArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PdocumentRemove
func P2PdocumentRemove(nodePtr C.uintptr_t, collections *C.char) *C.Result {
	ctx := context.Background()
	colArgs := splitCommaSeparatedString(C.GoString(collections))

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	err := node.Peer.RemoveP2PDocuments(ctx, colArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PdocumentGetAll
func P2PdocumentGetAll(nodePtr C.uintptr_t) *C.Result {
	ctx := context.Background()

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	cols, err := node.Peer.GetAllP2PDocuments(ctx)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(marshalJSONToGoCResult(cols))
}

//export P2PdocumentSync
func P2PdocumentSync(nodePtr C.uintptr_t, collection *C.char, docIDs *C.char, timeoutStr *C.char) *C.Result {
	ctx := context.Background()
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

	h := cgo.Handle(nodePtr)
	node := h.Value().(*node.Node)
	err := node.Peer.SyncDocuments(ctx, C.GoString(collection), docArgs)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
