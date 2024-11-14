// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"context"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/net"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
)

// p2pState contains all p2p related testing state.
type p2pState struct {
	// connections contains all connected nodes.
	//
	// The map key is the connected node id.
	connections map[int]struct{}

	// replicators is a mapping of replicator targets.
	//
	// The map key is the source node id.
	replicators map[int]struct{}

	// peerCollections contains all active peer collection subscriptions.
	//
	// The map key is the node id of the subscriber.
	peerCollections map[int]struct{}

	// actualDocHeads contains all document heads that exist on a node.
	//
	// The map key is the doc id. The map value is the doc head.
	actualDocHeads map[string]docHeadState

	// expectedDocHeads contains all document heads that are expected to exist on a node.
	//
	// The map key is the doc id. The map value is the doc head.
	expectedDocHeads map[string]cid.Cid
}

// docHeadState contains the state of a document head.
// It is used to track if a document at a certain head has been decrypted.
type docHeadState struct {
	// The actual document head.
	cid cid.Cid
	// Indicates if the document at the given head has been decrypted.
	decrypted bool
}

// newP2PState returns a new empty p2p state.
func newP2PState() *p2pState {
	return &p2pState{
		connections:      make(map[int]struct{}),
		replicators:      make(map[int]struct{}),
		peerCollections:  make(map[int]struct{}),
		actualDocHeads:   make(map[string]docHeadState),
		expectedDocHeads: make(map[string]cid.Cid),
	}
}

// eventState contains all event related testing state for a node.
type eventState struct {
	// merge is the `event.MergeCompleteName` subscription
	merge *event.Subscription

	// update is the `event.UpdateName` subscription
	update *event.Subscription

	// replicator is the `event.ReplicatorCompletedName` subscription
	replicator *event.Subscription

	// p2pTopic is the `event.P2PTopicCompletedName` subscription
	p2pTopic *event.Subscription
}

// newEventState returns an eventState with all required subscriptions.
func newEventState(bus *event.Bus) (*eventState, error) {
	merge, err := bus.Subscribe(event.MergeCompleteName)
	if err != nil {
		return nil, err
	}
	update, err := bus.Subscribe(event.UpdateName)
	if err != nil {
		return nil, err
	}
	replicator, err := bus.Subscribe(event.ReplicatorCompletedName)
	if err != nil {
		return nil, err
	}
	p2pTopic, err := bus.Subscribe(event.P2PTopicCompletedName)
	if err != nil {
		return nil, err
	}
	return &eventState{
		merge:      merge,
		update:     update,
		replicator: replicator,
		p2pTopic:   p2pTopic,
	}, nil
}

// nodeState contains all testing state for a node.
type nodeState struct {
	// The node's client active in this test.
	clients.Client
	// event contains all event node subscriptions.
	event *eventState
	// p2p contains p2p states for the node.
	p2p *p2pState
	// The network configurations for the nodes
	netOpts []net.NodeOpt
	// The path to any file-based databases active in this test.
	dbPath string
	// Collections by index present in the test.
	// Indexes matches that of collectionNames.
	collections []client.Collection
	// Indexes, by index, by collection index.
	indexes [][]client.IndexDescription
	// indicates if the node is closed.
	closed bool
	// peerInfo contains the peer information for the node.
	peerInfo peer.AddrInfo
}

// state contains all testing state.
type state struct {
	// The test context.
	ctx context.Context

	// The Go Test test state
	t testing.TB

	// The TestCase currently being executed.
	testCase TestCase

	// The type of KMS currently being tested.
	kms KMSType

	// The type of database currently being tested.
	dbt DatabaseType

	// The type of client currently being tested.
	clientType ClientType

	// Any explicit transactions active in this test.
	//
	// This is order dependent and the property is accessed by index.
	txns []datastore.Txn

	// identities contains all identities created in this test.
	// The map key is the identity reference that uniquely identifies identities of different
	// types. See [identRef].
	// The map value is the identity holder that contains the identity itself and token
	// generated for different target nodes. See [identityHolder].
	identities map[identityRef]*identityHolder

	// The seed for the next identity generation. We want identities to be deterministic.
	nextIdentityGenSeed int

	// Will receive an item once all actions have finished processing.
	allActionsDone chan struct{}

	// These channels will receive a function which asserts results of any subscription requests.
	subscriptionResultsChans []chan func()

	// The nodes active in this test.
	nodes []*nodeState

	// The ACP options to share between each node.
	acpOptions []node.ACPOpt

	// The names of the collections active in this test.
	// Indexes matches that of initial collections.
	collectionNames []string

	// A map of the collection indexes by their Root, this allows easier
	// identification of collections in a natural, human readable, order
	// even when they are renamed.
	collectionIndexesByRoot map[uint32]int

	// Document IDs by index, by collection index.
	//
	// Each index is assumed to be global, and may be expected across multiple
	// nodes.
	docIDs [][]client.DocID

	// Valid Cid string values by [UniqueCid] ID.
	cids map[any]string

	// isBench indicates wether the test is currently being benchmarked.
	isBench bool

	// The SourceHub address used to pay for SourceHub transactions.
	sourcehubAddress string

	// isNetworkEnabled indicates whether the network is enabled.
	isNetworkEnabled bool
}

// newState returns a new fresh state for the given testCase.
func newState(
	ctx context.Context,
	t testing.TB,
	testCase TestCase,
	kms KMSType,
	dbt DatabaseType,
	clientType ClientType,
	collectionNames []string,
) *state {
	return &state{
		ctx:                      ctx,
		t:                        t,
		testCase:                 testCase,
		kms:                      kms,
		dbt:                      dbt,
		clientType:               clientType,
		txns:                     []datastore.Txn{},
		allActionsDone:           make(chan struct{}),
		identities:               map[identityRef]*identityHolder{},
		subscriptionResultsChans: []chan func(){},
		collectionNames:          collectionNames,
		collectionIndexesByRoot:  map[uint32]int{},
		docIDs:                   [][]client.DocID{},
		cids:                     map[any]string{},
		isBench:                  false,
	}
}
