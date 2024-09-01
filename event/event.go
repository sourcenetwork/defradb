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
	"github.com/libp2p/go-libp2p/core/peer"
)

// Name identifies an event
type Name string

const (
	// WildCardName is the alias used to subscribe to all events.
	WildCardName = Name("*")
	// MergeName is the name of the net merge request event.
	MergeName = Name("merge")
	// MergeCompleteName is the name of the database merge complete event.
	MergeCompleteName = Name("merge-complete")
	// UpdateName is the name of the database update event.
	UpdateName = Name("update")
	// PubSubName is the name of the network pubsub event.
	PubSubName = Name("pubsub")
	// P2PTopicName is the name of the network p2p topic update event.
	P2PTopicName = Name("p2p-topic")
	// PeerInfoName is the name of the network peer info event.
	PeerInfoName = Name("peer-info")
	// ReplicatorName is the name of the replicator event.
	ReplicatorName = Name("replicator")
	// P2PTopicCompletedName is the name of the network p2p topic update completed event.
	P2PTopicCompletedName = Name("p2p-topic-completed")
	// ReplicatorCompletedName is the name of the replicator completed event.
	ReplicatorCompletedName = Name("replicator-completed")
)

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
	// DocID is the unique immutable identifier of the document that was updated.
	DocID string

	// ByPeer is the id of the peer that created the push log request.
	ByPeer peer.ID

	// FromPeer is the id of the peer that received the push log request.
	FromPeer peer.ID

	// Cid is the id of the composite commit that formed this update in the DAG.
	Cid cid.Cid

	// SchemaRoot is the root identifier of the schema that defined the shape of the document that was updated.
	SchemaRoot string
}

// Message contains event info.
type Message struct {
	// Name is the name of the event this message was generated from.
	Name Name

	// Data contains optional event information.
	Data any
}

// NewMessage returns a new message with the given name and optional data.
func NewMessage(name Name, data any) Message {
	return Message{name, data}
}

// Subscription is a read-only event stream.
type Subscription struct {
	id     uint64
	value  chan Message
	events []Name
}

// Message returns the next event value from the subscription.
func (s *Subscription) Message() <-chan Message {
	return s.value
}

// P2PTopic is an event that is published when a peer has updated the topics it is subscribed to.
type P2PTopic struct {
	ToAdd    []string
	ToRemove []string
}

// PeerInfo is an event that is published when the node has updated its peer info.
type PeerInfo struct {
	Info peer.AddrInfo
}

// Replicator is an event that is published when a replicator is added or updated.
type Replicator struct {
	// The peer info for the replicator instance.
	Info peer.AddrInfo
	// The map of schema roots that the replicator will receive updates for.
	Schemas map[string]struct{}
	// Docs will receive Updates if new collections have been added to the replicator
	// and those collections have documents to be replicated.
	Docs <-chan Update
}
