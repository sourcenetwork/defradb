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
	"strconv"
	"time"

	"github.com/sourcenetwork/defradb/internal/db"
	netConfig "github.com/sourcenetwork/defradb/net/config"
	"github.com/sourcenetwork/defradb/node"
)

//export NewNode
func NewNode(cOptions C.NodeInitOptions) C.NewNodeResult {
	gocOptions, err := convertNodeInitOptionsToGoNodeInitOptions(cOptions)
	if err != nil {
		return returnNewNodeResultC(1, err.Error(), nil)
	}

	inMemoryFlag := gocOptions.InMemory != 0
	listeningAddresses := splitCommaSeparatedString(gocOptions.ListeningAddresses)

	ctx := context.Background()

	// // Create the directory if it doesn't exist, and inMemory flag is not set
	// For now this is not done, but we leave it here because we might need it in
	// the future, when running on mobile platforms.
	// if !inMemoryFlag {
	// 	if _, err = os.Stat(gocOptions.DbPath); os.IsNotExist(err) {
	// 		err := os.MkdirAll(gocOptions.DbPath, 0755)
	// 		if err != nil {
	// 			return returnGoC(1, err.Error(), "")
	// 		}
	// 	}
	// }

	opts := []node.Option{
		node.WithStorePath(gocOptions.DbPath),
		node.WithLensRuntime(node.Wazero),
	}
	if len(listeningAddresses) > 0 {
		opts = append(opts, netConfig.WithListenAddresses(listeningAddresses...))
	}
	maxTxnRetries := gocOptions.MaxTransactionRetries
	if maxTxnRetries > 0 {
		opts = append(opts, db.WithMaxRetries(maxTxnRetries))
	}
	disableP2PFlag := gocOptions.DisableP2P != 0
	if disableP2PFlag {
		opts = append(opts, node.WithDisableP2P(true))
	}
	disableAPIFlag := gocOptions.DisableAPI != 0
	if disableAPIFlag {
		opts = append(opts, node.WithDisableAPI(true))
	}
	if inMemoryFlag {
		opts = append(opts, node.WithBadgerInMemory(true))
	}
	peers := splitCommaSeparatedString(gocOptions.Peers)
	if len(peers) > 0 {
		opts = append(opts, netConfig.WithBootstrapPeers(peers...))
	}
	if gocOptions.Identity != nil {
		opts = append(opts, db.WithNodeIdentity(gocOptions.Identity))
	}
	if gocOptions.EnableNodeACP != 0 {
		opts = append(opts, node.WithEnableNodeACP(true))
	}
	opts = append(opts, node.WithDocumentACPPath(""))
	opts = append(opts, node.WithNodeACPPath(""))

	// Configure the replicator retry times. Go from string slice -> time.Duration slice
	replicatorRetryTimes := splitCommaSeparatedString(gocOptions.ReplicatorRetryIntervals)
	var replicatorRetryIntervals []time.Duration
	for _, s := range replicatorRetryTimes {
		n, err := strconv.Atoi(s)
		if err != nil {
			return returnNewNodeResultC(1, err.Error(), nil)
		}
		if n <= 0 {
			return returnNewNodeResultC(1, errNegativeReplicatorTime, nil)
		}
		replicatorRetryIntervals = append(replicatorRetryIntervals, time.Duration(n)*time.Second)
	}
	if len(replicatorRetryIntervals) > 0 {
		opts = append(opts, db.WithRetryInterval(replicatorRetryIntervals))
	}

	n, err := node.New(ctx, opts...)
	if err != nil {
		return returnNewNodeResultC(1, err.Error(), nil)
	}
	err = n.Start(ctx)
	if err != nil {
		return returnNewNodeResultC(1, err.Error(), nil)
	}
	return returnNewNodeResultC(0, "", n)
}

//export NodeClose
func NodeClose(nodePtr C.uintptr_t) C.Result {
	node, err := getNodeFromPointer(nodePtr)
	if err != nil {
		return returnC(returnGoC(1, err.Error(), ""))
	}
	err = node.Close(context.Background())
	if err != nil {
		return returnC(GoCResult{1, err.Error(), ""})
	}
	return returnC(GoCResult{0, "", ""})
}
