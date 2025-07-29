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
	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/immutable"
)

// Bus handles routing and publishing of messages to subscribers.
type Bus interface {
	// Publish broadcasts the given message to the bus subscribers. Non-blocking.
	Publish(msg Message)
	// Subscribe returns a new subscription that will receive all of the events
	// contained in the given list of events.
	Subscribe(events ...Name) (Subscription, error)
	// Unsubscribe removes all event subscriptions and closes the subscription.
	//
	// Will do nothing if this object is already closed.
	Unsubscribe(sub Subscription)
	// Close unsubscribes all active subscribers and closes the bus.
	Close()
}

// Subscription receives subscribed messages until closed.
type Subscription interface {
	// Message returns the message channel for the subscription.
	Message() <-chan Message
}

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
	// PeerInfoName is the name of the network peer info event.
	PeerInfoName = Name("peer-info")
	// ReplicatorName is the name of the replicator event.
	ReplicatorName = Name("replicator")
	// ReplicatorFailureName is the name of the replicator failure event.
	ReplicatorFailureName = Name("replicator-failure")
	// ReplicatorCompletedName is the name of the replicator completed event.
	ReplicatorCompletedName = Name("replicator-completed")
	// PurgeName is the name of the purge event.
	PurgeName = Name("purge")
	// DocUpdateRequestName is the name of the document update request event.
	DocUpdateRequestName = Name("doc-update-request")
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

	// CollectionID is the root identifier of the collection that this document goes by.
	CollectionID string

	// Block is the encoded contents of this composite commit, it contains the Cids of the field level commits that
	// also formed this update.
	Block []byte

	// IsRetry is true if this update is a retry of a previously failed update.
	IsRetry bool

	// Identity is the identity of the peer that created this update.
	Identity immutable.Option[acpIdentity.Identity]
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

	// CollectionID is the root identifier of the collection that this document goes by.
	CollectionID string
}

// MergeComplete is a notification that a merge has been completed.
type MergeComplete struct {
	// Merge is the merge that was completed.
	Merge Merge

	// Decrypted specifies if the merge payload was decrypted.
	Decrypted bool
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

// ReplicatorFailure is an event that is published when a replicator fails to replicate a document.
type ReplicatorFailure struct {
	// PeerID is the id of the peer that failed to replicate the document.
	PeerID peer.ID
	// DocID is the unique immutable identifier of the document that failed to replicate.
	DocID string
}

// DocUpdateRequest is an event that is published when a node needs to request
// a specific document from the network.
type DocUpdateRequest struct {
	// CollectionID is the collection identifier.
	CollectionID string
	// DocID is the document identifier to request.
	DocID string
	// Response is a channel to receive the response.
	Response chan DocUpdateResponse
}

// DocUpdateResponse is the response to a DocUpdateRequest.
type DocUpdateResponse struct {
	// Found indicates if the document was found on any peer.
	Found bool
	// Error contains any error that occurred during the request.
	Error error
}
