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
	"time"

	"github.com/sourcenetwork/defradb/client"
)

//export P2PInfo
func P2PInfo(nodePtr C.uintptr_t) C.Result {
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	info := node.DB.PeerInfo()
	return returnC(marshalJSONToGoCResult(info))
}

//export P2PgetAllReplicators
func P2PgetAllReplicators(nodePtr C.uintptr_t) C.Result {
	ctx := context.Background()
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
func P2PsetReplicator(nodePtr C.uintptr_t, collections *C.char, peerInfo *C.char) C.Result {
	ctx := context.Background()
	colArgs := splitCommaSeparatedString(C.GoString(collections))

	var info client.PeerInfo
	if err := json.Unmarshal([]byte(C.GoString(peerInfo)), &info); err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	err = node.DB.SetReplicator(ctx, info, colArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PdeleteReplicator
func P2PdeleteReplicator(nodePtr C.uintptr_t, collections *C.char, peerInfo *C.char) C.Result {
	ctx := context.Background()
	colArgs := splitCommaSeparatedString(C.GoString(collections))

	var info client.PeerInfo
	if err := json.Unmarshal([]byte(C.GoString(peerInfo)), &info); err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}

	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	err = node.DB.DeleteReplicator(ctx, info, colArgs...)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}

//export P2PcollectionAdd
func P2PcollectionAdd(nodePtr C.uintptr_t, collections *C.char) C.Result {
	ctx := context.Background()
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
func P2PcollectionRemove(nodePtr C.uintptr_t, collections *C.char) C.Result {
	ctx := context.Background()
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
func P2PcollectionGetAll(nodePtr C.uintptr_t) C.Result {
	ctx := context.Background()

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
func P2PdocumentAdd(nodePtr C.uintptr_t, collections *C.char) C.Result {
	ctx := context.Background()
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
func P2PdocumentRemove(nodePtr C.uintptr_t, collections *C.char) C.Result {
	ctx := context.Background()
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
func P2PdocumentGetAll(nodePtr C.uintptr_t) C.Result {
	ctx := context.Background()
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
func P2PdocumentSync(nodePtr C.uintptr_t, collection *C.char, docIDs *C.char, timeoutStr *C.char) C.Result {
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

//export P2Pconnect
func P2Pconnect(nodePtr C.uintptr_t, peerID *C.char, peerAddresses *C.char) C.Result {
	ctx := context.Background()
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	var info client.PeerInfo
	info.ID = C.GoString(peerID)
	info.Addresses = splitCommaSeparatedString(C.GoString(peerAddresses))
	err = node.DB.Connect(ctx, info)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	return returnC(returnGoC(0, "", ""))
}
