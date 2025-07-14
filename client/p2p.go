// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"context"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
)

// P2P is a peer connected database implementation.
type P2P interface {
	// PeerInfo returns the p2p host id and listening addresses.
	PeerInfo() peer.AddrInfo

	// SetReplicator adds a replicator to the persisted list or adds
	// schemas if the replicator already exists.
	SetReplicator(ctx context.Context, info peer.AddrInfo, collectionNames ...string) error
	// DeleteReplicator deletes a replicator from the persisted list
	// or specific schemas if they are specified.
	DeleteReplicator(ctx context.Context, info peer.AddrInfo, collectionNames ...string) error
	// GetAllReplicators returns the full list of replicators with their
	// subscribed schemas.
	GetAllReplicators(ctx context.Context) ([]Replicator, error)

	// AddP2PCollections adds the given collection IDs to the P2P system and
	// subscribes to their topics. It will error if any of the provided
	// collection IDs are invalid.
	AddP2PCollections(ctx context.Context, collectionIDs ...string) error

	// RemoveP2PCollections removes the given collection IDs from the P2P system and
	// unsubscribes from their topics. It will error if the provided
	// collection IDs are invalid.
	RemoveP2PCollections(ctx context.Context, collectionIDs ...string) error

	// GetAllP2PCollections returns the list of persisted collection IDs that
	// the P2P system subscribes to.
	GetAllP2PCollections(ctx context.Context) ([]string, error)

	// SyncDocuments requests the latest versions of specified documents from the network
	// and synchronizes their DAGs locally. After successful sync, automatically subscribes
	// to the documents and their collection for future updates.
	//
	// Returns a map of document ID to sync result with head CIDs.
	SyncDocuments(ctx context.Context, collectionID string, docIDs []string, opts ...DocSyncOption) <-chan error
}

// DocSyncOption configures the document sync operation.
type DocSyncOption func(*DocSyncOptions)

// DocSyncOptions contains options for document sync operations.
type DocSyncOptions struct {
	Timeout time.Duration
}

// DocSyncWithTimeout sets the timeout for the sync operation.
func DocSyncWithTimeout(timeout time.Duration) DocSyncOption {
	return func(opts *DocSyncOptions) {
		opts.Timeout = timeout
	}
}

// DocSyncResult represents the result of synchronizing a single document.
type DocSyncResult struct {
	// Heads is the list of CID heads of the document.
	Heads []cid.Cid
	// Sender is the ID of the peer that provided the document.
	Sender string
}
