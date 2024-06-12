// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package event

import (
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	// WildCardEventName is the alias used to subscribe to all events.
	WildCardEventName = "*"
	// MergeEventName is the name of the net merge request event.
	MergeEventName = "merge"
	// MergeCompleteEventName is the name of the database merge complete event.
	MergeCompleteEventName = "merge-complete"
	// UpdateEventName is the name of the database update event.
	UpdateEventName = "update"
	// PubSubEventName is the name of the network pubsub event.
	PubSubEventName = "pubsub"
	// PeerEventName is the name of the network connect event.
	PeerEventName = "peer"
)

// Peer is an event that is published when
// a peer connection has changed status.
type Peer = event.EvtPeerConnectednessChanged

// PubSub is an event that is published when
// a pubsub message has been received from a remote peer.
type PubSub struct {
	// Peer is the id of the peer that published the message.
	Peer peer.ID
}

// Update represents a new DAG node added to the append-only composite MerkleCRDT Clock graph
// of a document.
//
// It must only contain public elements not protected by ACP.
type Update struct {
	// DocID is the unique immutable identifier of the document that was updated.
	DocID string

	// Cid is the id of the composite commit that formed this update in the DAG.
	Cid cid.Cid

	// SchemaRoot is the root identifier of the schema that defined the shape of the document that was updated.
	SchemaRoot string

	// Block is the encoded contents of this composite commit, it contains the Cids of the field level commits that
	// also formed this update.
	Block []byte

	// IsCreate is true if this update is the creation of a new document.
	IsCreate bool
}

// Merge is a notification that a merge can be performed up to the provided CID.
type Merge struct {
	// ByPeer is the id of the peer that created the push log request.
	ByPeer peer.ID

	// FromPeer is the id of the peer that received the push log request.
	FromPeer peer.ID

	// Cid is the id of the composite commit that formed this update in the DAG.
	Cid cid.Cid

	// SchemaRoot is the root identifier of the schema that defined the shape of the document that was updated.
	SchemaRoot string
}
