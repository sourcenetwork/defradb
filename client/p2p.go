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
	"encoding/json"
	"io"

	"github.com/ipfs/go-cid"

	"github.com/sourcenetwork/corekv/blockstore"
)

// P2P is a peer connected database implementation.
type P2P interface {
	// PeerInfo returns the p2p host list of addresses.
	PeerInfo() ([]string, error)

	// Connect tries to connect to the peer with the given [PeerInfo].
	Connect(ctx context.Context, addresses []string) error

	// SetReplicator adds a replicator to the persisted list or adds
	// schemas if the replicator already exists.
	SetReplicator(ctx context.Context, addresses []string, collectionNames ...string) error
	// DeleteReplicator deletes a replicator from the persisted list
	// or specific schemas if they are specified.
	DeleteReplicator(ctx context.Context, id string, collectionNames ...string) error
	// GetAllReplicators returns the full list of replicators with their
	// subscribed schemas.
	GetAllReplicators(ctx context.Context) ([]Replicator, error)

	// AddP2PCollections adds the given collections to the P2P system and
	// subscribes to their topics. It will error if any of the provided
	// collection names are invalid.
	AddP2PCollections(ctx context.Context, collectionNames ...string) error

	// RemoveP2PCollections removes the given collections from the P2P system and
	// unsubscribes from their topics. It will error if the provided
	// collection names are invalid.
	RemoveP2PCollections(ctx context.Context, collectionNames ...string) error

	// GetAllP2PCollections returns the list of persisted collection names that
	// the P2P system subscribes to.
	GetAllP2PCollections(ctx context.Context) ([]string, error)

	// AddP2PDocuments adds the given docIDs to the P2P system and
	// subscribes to their topics. It will error if any of the provided
	// docIDs are invalid.
	AddP2PDocuments(ctx context.Context, docIDs ...string) error

	// RemoveP2PDocuments removes the given docIDs from the P2P system and
	// unsubscribes from their topics. It will error if the provided
	// docIDs are invalid.
	RemoveP2PDocuments(ctx context.Context, docIDs ...string) error

	// GetAllP2PDocuments returns the list of persisted docIDs that
	// the P2P system subscribes to.
	GetAllP2PDocuments(ctx context.Context) ([]string, error)

	// SyncDocuments requests the latest versions of specified documents from the network
	// and synchronizes their DAGs locally. It doesn't automatically subscribe
	// to the documents or their collection for future updates.
	// context.WithTimeout can be used to set a timeout for the operation.
	SyncDocuments(ctx context.Context, collectionName string, docIDs []string) error

	// SyncCollectionVersions synchronizes the given collection versions to local node.
	//
	// It will not complete until a version is found, so it is strongly recommended
	// to set a timeout using `context.WithTimeout`.
	SyncCollectionVersions(ctx context.Context, versionIDs ...string) error
}

type StreamHandler = func(stream io.Reader, peerID string)
type PubsubMessageHandler = func(from string, topic string, msg []byte) ([]byte, error)
type BlockAccessFunc = func(ctx context.Context, peerID string, c cid.Cid) bool

type PeerInfo struct {
	ID        string
	Addresses []string
}

func (p PeerInfo) String() string {
	b, _ := json.Marshal(p)
	return string(b)
}

type PubsubResponse = struct {
	// ID is the cid.Cid of the received message.
	ID string
	// From is the ID of the sender.
	From string
	// Data is the message data.
	Data []byte
	// Err is an error from the sender.
	Err error
}

type Host interface {
	// ID returns the peer ID of the host.
	ID() string
	// Addrs returns the host's list of addresses.
	Addresses() ([]string, error)
	// Pubkey return the byte slice representation of the host's public key.
	Pubkey() ([]byte, error)
	// Connect tries to connect to the peer with the given addresses.
	Connect(ctx context.Context, addresses []string) error
	// Disconnect will try to disconnect from the peer with the given ID.
	Disconnect(ctx context.Context, peerID string) error
	// Send will try to send the given data to a peer.
	Send(ctx context.Context, data []byte, peerID string, protocolID string) error
	// Sign will return a hash of the provided data signed with the private key of the host.
	Sign(data []byte) ([]byte, error)
	// SetStreamHandler tells the host to listen for messages of the provided protocol ID and
	// handle them with the given handler.
	SetStreamHandler(protocolID string, handler StreamHandler)
	// AddPubSubTopic adds a pubsub topic to the host.
	AddPubSubTopic(topicName string, subscribe bool, handler PubsubMessageHandler) error
	// RemovePubSubTopic removes the given topic from the host.
	RemovePubSubTopic(topic string) error
	// PublishToTopicAsync sends a new message on the given topic without waiting for a response.
	PublishToTopicAsync(ctx context.Context, topic string, data []byte) error
	// PublishToTopic sends a new message on the given topic, returning a response channel.
	// It provides the option to allow responses from multiple peers.
	//
	// NOTE: The returned channel type is leaking from the go-p2p package so its not ideal. We should
	// consider finding a better solution.
	PublishToTopic(
		ctx context.Context,
		topic string,
		data []byte,
		withMultiResponse bool,
	) (<-chan PubsubResponse, error)
	// IPLDStore returns the host's IPLD store implementation.
	IPLDStore() blockstore.IPLDStore
	// ContextWithSession returns a new context with a session for the underlying block service..
	ContextWithSession(ctx context.Context) context.Context
	// SetBlockAccessFunc set the function to use to determine if a peer has access to
	// the requested blocks on the block service.
	SetBlockAccessFunc(accessFunc BlockAccessFunc)
}
