// Copyright 2026 Democratized Data Foundation
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

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client/options"
	iIdentity "github.com/sourcenetwork/defradb/internal/identity"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/immutable"
)

//export NewNode
func NewNode(cOptions C.NodeInitOptions) C.NewNodeResult {
	gocOptions, err := convertNodeInitOptionsToGoNodeInitOptions(cOptions)
	if err != nil {
		return returnNewNodeResultC(1, err.Error(), nil)
	}

	ctx := context.Background()
	opts := options.Node()

	// Node-level options
	if gocOptions.DisableP2P != 0 {
		opts.SetDisableP2P(true)
	}
	if gocOptions.DisableAPI != 0 {
		opts.SetDisableAPI(true)
	}

	// Store options
	if gocOptions.StoreType != "" {
		opts.Store().SetType(options.NodeStoreType(gocOptions.StoreType))
	}
	if gocOptions.DbPath != "" {
		opts.Store().SetPath(gocOptions.DbPath)
	}
	if gocOptions.InMemory != 0 {
		opts.Store().SetBadgerInMemory(true)
	}
	if gocOptions.BadgerFileSize > 0 {
		opts.Store().SetBadgerFileSize(gocOptions.BadgerFileSize)
	}
	if len(gocOptions.BadgerEncryptionKey) > 0 {
		opts.Store().SetBadgerEncryptionKey(gocOptions.BadgerEncryptionKey)
	}

	// DB options
	// Currently the only supported lens runtime is wazero, so use it explicitly
	opts.DB().SetLensRuntime("wazero")
	if gocOptions.MaxTransactionRetries > 0 {
		opts.DB().SetMaxTxnRetries(gocOptions.MaxTransactionRetries)
	}
	if gocOptions.Identity != nil {
		ctx = iIdentity.WithContext(ctx, immutable.Some[acpIdentity.Identity](gocOptions.Identity))
		opts.DB().SetNodeIdentity(gocOptions.Identity)
	}
	if gocOptions.EnableSigning != 0 {
		opts.DB().SetEnableSigning(true)
	}
	if len(gocOptions.SearchableEncryptionKey) > 0 {
		opts.DB().SetSearchableEncryptionKey(gocOptions.SearchableEncryptionKey)
	}
	if gocOptions.P2PBlockSyncTimeoutMs > 0 {
		opts.DB().SetP2PBlockSyncTimeout(time.Duration(gocOptions.P2PBlockSyncTimeoutMs) * time.Millisecond)
	}
	if gocOptions.LensPoolSize > 0 {
		opts.DB().SetLensPoolSize(gocOptions.LensPoolSize)
	}
	if gocOptions.ChunkSize > 0 {
		opts.DB().SetChunkSize(gocOptions.ChunkSize)
	}
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
		opts.DB().SetRetryIntervals(replicatorRetryIntervals)
	}

	// P2P options
	listeningAddresses := splitCommaSeparatedString(gocOptions.ListeningAddresses)
	if len(listeningAddresses) > 0 {
		opts.P2P().SetListenAddresses(listeningAddresses...)
	}
	peers := splitCommaSeparatedString(gocOptions.Peers)
	if len(peers) > 0 {
		opts.P2P().SetBootstrapPeers(peers...)
	}
	if gocOptions.EnablePubSub != 0 {
		opts.P2P().SetEnablePubSub(true)
	}
	if gocOptions.EnableRelay != 0 {
		opts.P2P().SetEnableRelay(true)
	}
	if gocOptions.EnableClearBackoffOnRetry != 0 {
		opts.P2P().SetEnableClearBackoffOnRetry(true)
	}
	if len(gocOptions.P2PPrivateKey) > 0 {
		opts.P2P().SetPrivateKey(gocOptions.P2PPrivateKey)
	}

	// HTTP options
	if gocOptions.HTTPAddress != "" {
		opts.HTTP().SetAddress(gocOptions.HTTPAddress)
	}
	allowedOrigins := splitCommaSeparatedString(gocOptions.HTTPAllowedOrigins)
	if len(allowedOrigins) > 0 {
		opts.HTTP().SetAllowedOrigins(allowedOrigins...)
	}
	if gocOptions.TLSCertPath != "" {
		opts.HTTP().SetCertPath(gocOptions.TLSCertPath)
	}
	if gocOptions.TLSKeyPath != "" {
		opts.HTTP().SetKeyPath(gocOptions.TLSKeyPath)
	}
	if gocOptions.HTTPReadTimeoutMs > 0 {
		opts.HTTP().SetReadTimeout(time.Duration(gocOptions.HTTPReadTimeoutMs) * time.Millisecond)
	}
	if gocOptions.HTTPWriteTimeoutMs > 0 {
		opts.HTTP().SetWriteTimeout(time.Duration(gocOptions.HTTPWriteTimeoutMs) * time.Millisecond)
	}
	if gocOptions.HTTPIdleTimeoutMs > 0 {
		opts.HTTP().SetIdleTimeout(time.Duration(gocOptions.HTTPIdleTimeoutMs) * time.Millisecond)
	}

	// Document ACP options
	if gocOptions.DocumentACPType != "" {
		opts.DocumentACP().SetType(options.NodeDocumentACPType(gocOptions.DocumentACPType))
	}
	if gocOptions.DocumentACPPath != "" {
		opts.DocumentACP().SetPath(gocOptions.DocumentACPPath)
	}
	if gocOptions.RemoteDACLogID != "" {
		opts.DocumentACP().SetLogID(gocOptions.RemoteDACLogID)
	}
	if gocOptions.RemoteDACGRPCAddress != "" {
		opts.DocumentACP().SetGRPCAddress(gocOptions.RemoteDACGRPCAddress)
	}
	if gocOptions.RemoteDACCometRPCAddress != "" {
		opts.DocumentACP().SetCometRPCAddress(gocOptions.RemoteDACCometRPCAddress)
	}

	// Node ACP options
	if gocOptions.EnableNodeACP != 0 {
		opts.NodeACP().SetEnabled(true)
	}
	if gocOptions.NodeACPPath != "" {
		opts.NodeACP().SetPath(gocOptions.NodeACPPath)
	}

	n, err := node.New(ctx, opts)
	if err != nil {
		return returnNewNodeResultC(1, err.Error(), nil)
	}
	err = n.Start(ctx)
	if err != nil {
		return returnNewNodeResultC(1, err.Error(), nil)
	}
	return returnNewNodeResultC(0, "", n)
}
