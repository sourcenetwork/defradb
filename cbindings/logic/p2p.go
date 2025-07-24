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

import (
	"context"
	"encoding/json"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

func P2PInfo() GoCResult {
	info := globalNode.Peer.PeerInfo()
	return marshalJSONToGoCResult(info)
}

func P2PgetAllReplicators() GoCResult {
	ctx := context.Background()
	reps, err := globalNode.Peer.GetAllReplicators(ctx)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return marshalJSONToGoCResult(reps)
}

func P2PsetReplicator(collections string, peerStr string, txnID uint64) GoCResult {
	ctx := context.Background()
	colArgs := splitCommaSeparatedString(collections)

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	var info peer.AddrInfo
	if err := json.Unmarshal([]byte(peerStr), &info); err != nil {
		return returnGoC(1, err.Error(), "")
	}

	err = globalNode.Peer.SetReplicator(ctx, info, colArgs...)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}

func P2PdeleteReplicator(collections string, peerStr string, txnID uint64) GoCResult {
	ctx := context.Background()
	colArgs := splitCommaSeparatedString(collections)

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	var info peer.AddrInfo
	if err := json.Unmarshal([]byte(peerStr), &info); err != nil {
		return returnGoC(1, err.Error(), "")
	}

	err = globalNode.Peer.DeleteReplicator(ctx, info, colArgs...)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}

func P2PcollectionAdd(collections string, txnID uint64) GoCResult {
	ctx := context.Background()
	colArgs := splitCommaSeparatedString(collections)

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	err = globalNode.Peer.AddP2PCollections(ctx, colArgs...)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}

func P2PcollectionRemove(collections string, txnID uint64) GoCResult {
	ctx := context.Background()
	colArgs := splitCommaSeparatedString(collections)

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	err = globalNode.Peer.RemoveP2PCollections(ctx, colArgs...)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}

func P2PcollectionGetAll(txnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	cols, err := globalNode.Peer.GetAllP2PCollections(ctx)

	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return marshalJSONToGoCResult(cols)
}

func P2PdocumentAdd(collections string, txnID uint64) GoCResult {
	ctx := context.Background()
	colArgs := splitCommaSeparatedString(collections)

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	err = globalNode.Peer.AddP2PDocuments(ctx, colArgs...)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}

func P2PdocumentRemove(collections string, txnID uint64) GoCResult {
	ctx := context.Background()
	colArgs := splitCommaSeparatedString(collections)

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	err = globalNode.Peer.RemoveP2PDocuments(ctx, colArgs...)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}

func P2PdocumentGetAll(txnID uint64) GoCResult {
	ctx := context.Background()

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	cols, err := globalNode.Peer.GetAllP2PDocuments(ctx)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return marshalJSONToGoCResult(cols)
}

func P2PdocumentSync(collection string, docIDs string, txnID uint64, timeout string) GoCResult {
	ctx := context.Background()
	docArgs := splitCommaSeparatedString(docIDs)
	timeoutDuration := time.Duration(0)

	if timeout != "" {
		timeoutDurationParsed, err := time.ParseDuration(timeout)
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		timeoutDuration = timeoutDurationParsed
	}

	ctx, err := contextWithTransaction(ctx, txnID)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	if timeoutDuration > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeoutDuration)
		defer cancel()
	}

	err = globalNode.Peer.SyncDocuments(ctx, collection, docArgs)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}
	return returnGoC(0, "", "")
}
