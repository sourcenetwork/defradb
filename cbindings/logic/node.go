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
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/node"

	netConfig "github.com/sourcenetwork/defradb/net/config"
)

var (
	globalNodes   = make(map[int]*node.Node)
	globalNodesMu sync.RWMutex
)

func NodeInit(n int, cOptions GoNodeInitOptions) GoCResult {
	var err error

	inMemoryFlag := cOptions.InMemory != 0
	listeningAddresses := splitCommaSeparatedString(cOptions.ListeningAddresses)

	nodeIdentity, err := identityFromKey(cOptions.IdentityKeyType, cOptions.IdentityPrivateKey)
	if err != nil {
		return returnGoC(1, err.Error(), "")
	}

	ctx := context.Background()

	globalNodesMu.Lock()
	defer globalNodesMu.Unlock()

	if globalNodes[n] != nil {
		err := globalNodes[n].Close(ctx)
		if err != nil {
			return returnGoC(1, fmt.Sprintf(errClosingNode, err), "")
		}
		globalNodes[n] = nil
	}

	// Create the directory if it doesn't exist, and inMemory flag is not set
	if !inMemoryFlag {
		if _, err = os.Stat(cOptions.DbPath); os.IsNotExist(err) {
			err := os.MkdirAll(cOptions.DbPath, 0755)
			if err != nil {
				return returnGoC(1, fmt.Sprintf(errCreatingStoreDirectory, err), "")
			}
		}
	}

	opts := []node.Option{
		node.WithStorePath(cOptions.DbPath),
		node.WithLensRuntime(node.Wazero),
	}
	if len(listeningAddresses) > 0 {
		opts = append(opts, netConfig.WithListenAddresses(listeningAddresses...))
	}
	maxTxnRetries := int(cOptions.MaxTransactionRetries)
	if maxTxnRetries > 0 {
		opts = append(opts, db.WithMaxRetries(maxTxnRetries))
	}
	disableP2PFlag := cOptions.DisableP2P != 0
	if disableP2PFlag {
		opts = append(opts, node.WithDisableP2P(true))
	}
	disableAPIFlag := cOptions.DisableAPI != 0
	if disableAPIFlag {
		opts = append(opts, node.WithDisableAPI(true))
	}
	if inMemoryFlag {
		opts = append(opts, node.WithBadgerInMemory(true))
	}
	peers := splitCommaSeparatedString(cOptions.Peers)
	if len(peers) > 0 {
		opts = append(opts, netConfig.WithBootstrapPeers(peers...))
	}
	if nodeIdentity != nil {
		opts = append(opts, db.WithNodeIdentity(nodeIdentity))
	}
	if cOptions.EnableNodeACP != 0 {
		opts = append(opts, node.WithEnableNodeACP(true))
	}
	opts = append(opts, node.WithDocumentACPPath(""))
	opts = append(opts, node.WithNodeACPPath(""))

	// Configure the replicator retry times. Go from string slice -> time.Duration slice
	replicatorRetryTimes := splitCommaSeparatedString(cOptions.ReplicatorRetryIntervals)
	var replicatorRetryIntervals []time.Duration
	for _, s := range replicatorRetryTimes {
		n, err := strconv.Atoi(s)
		if err != nil {
			return returnGoC(1, fmt.Sprintf(errParsingReplicatorTimes, err), "")
		}
		if n <= 0 {
			return returnGoC(1, errNegativeReplicatorTime, "")
		}
		replicatorRetryIntervals = append(replicatorRetryIntervals, time.Duration(n)*time.Second)
	}
	if len(replicatorRetryIntervals) > 0 {
		opts = append(opts, netConfig.WithRetryInterval(replicatorRetryIntervals))
	}

	globalNodes[n], err = node.New(ctx, opts...)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(errCreatingNode, err), "")
	}

	return returnGoC(0, "", "")
}

func NodeStart(n int) GoCResult {
	globalNodesMu.Lock()
	defer globalNodesMu.Unlock()

	if globalNodes[n] == nil {
		return returnGoC(1, errUninitializedNode, "")
	}
	ctx := context.Background()

	errCh := make(chan error, 1)

	go func() {
		err := globalNodes[n].Start(ctx)
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return returnGoC(1, err.Error(), "")
		}
		return returnGoC(0, "", "")
	case <-time.After(5 * time.Second):
		// Timeout occurred, node may still start later
		return returnGoC(2, errUnreadyStart, "")
	}
}

func NodeStop(n int) GoCResult {
	globalNodesMu.Lock()
	defer globalNodesMu.Unlock()

	if globalNodes[n] == nil {
		return returnGoC(1, errStoppedNode, "")
	}
	ctx := context.Background()
	err := globalNodes[n].Close(ctx)
	if err != nil {
		return returnGoC(1, fmt.Sprintf(errStoppingNode, err), "")
	}
	globalNodes[n] = nil

	return returnGoC(0, "", "")
}

// GetNode is a thread-safe getter for a global node
// It is exported so that it can be used for integration testing
func GetNode(n int) *node.Node {
	globalNodesMu.RLock()
	defer globalNodesMu.RUnlock()
	return globalNodes[n]
}
