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

	"github.com/libp2p/go-libp2p/core/peer"
)

// P2P is a peer connected database implementation.
type P2P interface {
	DB

	// PeerInfo returns the p2p host id and listening addresses.
	PeerInfo() peer.AddrInfo

	// SetReplicator adds a replicator to the persisted list or adds
	// schemas if the replicator already exists.
	SetReplicator(ctx context.Context, rep Replicator) error
	// DeleteReplicator deletes a replicator from the persisted list
	// or specific schemas if they are specified.
	DeleteReplicator(ctx context.Context, rep Replicator) error
	// GetAllReplicators returns the full list of replicators with their
	// subscribed schemas.
	GetAllReplicators(ctx context.Context) ([]Replicator, error)

	// AddP2PCollection adds the given collection ID that the P2P system
	// subscribes to to the the persisted list. It will error if the provided
	// collection ID is invalid.
	AddP2PCollection(ctx context.Context, collectionID string) error

	// RemoveP2PCollection removes the given collection ID that the P2P system
	// subscribes to from the the persisted list. It will error if the provided
	// collection ID is invalid.
	RemoveP2PCollection(ctx context.Context, collectionID string) error

	// GetAllP2PCollections returns the list of persisted collection IDs that
	// the P2P system subscribes to.
	GetAllP2PCollections(ctx context.Context) ([]string, error)
}
