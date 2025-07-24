// Copyright 2025 Democratized Data Foundation
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

	"github.com/sourcenetwork/immutable"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/event"
	netConfig "github.com/sourcenetwork/defradb/net/config"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
)

// P2pState contains all p2p related testing state.
type P2pState struct {
	// Connections contains all connected nodes.
	//
	// The map key is the connected node id.
	Connections map[int]struct{}

	// Replicators is a mapping of replicator targets.
	//
	// The map key is the source node id.
	Replicators map[int]struct{}

	// PeerCollections contains all active peer collection subscriptions.
	//
	// The map key is the node id of the subscriber.
	PeerCollections map[int]struct{}

	// PeerDocuments contains all active peer document subscriptions.
	//
	// The map key is the node id of the subscriber.
	PeerDocuments map[ColDocIndex]struct{}

	// ActualDAGHeads contains all DAG heads that exist on a node.
	//
	// The map key is the doc id. The map value is the doc head.
	//
	// This tracks composite commits for documents, and collection commits for
	// branchable collections
	ActualDAGHeads map[string]DocHeadState

	// ExpectedDAGHeads contains all DAG heads that are expected to exist on a node.
	//
	// The map key is the doc id. The map value is the DAG head.
	//
	// This tracks composite commits for documents, and collection commits for
	// branchable collections
	ExpectedDAGHeads map[string]cid.Cid
}

// DocHeadState contains the state of a document head.
// It is used to track if a document at a certain head has been decrypted.
type DocHeadState struct {
	// The actual document head.
	Cid cid.Cid
	// Indicates if the document at the given head has been Decrypted.
	Decrypted bool
}

// NewP2PState returns a new empty p2p state.
func NewP2PState() *P2pState {
	return &P2pState{
		Connections:      make(map[int]struct{}),
		Replicators:      make(map[int]struct{}),
		PeerCollections:  make(map[int]struct{}),
		PeerDocuments:    make(map[ColDocIndex]struct{}),
		ActualDAGHeads:   make(map[string]DocHeadState),
		ExpectedDAGHeads: make(map[string]cid.Cid),
	}
}

// EventState contains all event related testing state for a node.
type EventState struct {
	// Merge is the `event.MergeCompleteName` subscription
	Merge event.Subscription

	// Update is the `event.UpdateName` subscription
	Update event.Subscription

	// Replicator is the `event.ReplicatorCompletedName` subscription
	Replicator event.Subscription
}

// NewEventState returns an eventState with all required subscriptions.
func NewEventState(bus event.Bus) (*EventState, error) {
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
	return &EventState{
		Merge:      merge,
		Update:     update,
		Replicator: replicator,
	}, nil
}

// NodeState contains all testing state for a node.
type NodeState struct {
	// The node's client active in this test.
	clients.Client
	// Event contains all Event node subscriptions.
	Event *EventState
	// P2p contains P2p states for the node.
	P2p *P2pState
	// The network configurations for the nodes
	NetOpts []netConfig.NodeOpt
	// The path to any file-based databases active in this test.
	DbPath string
	// Collections by index present in the test.
	// Indexes matches that of collectionNames.
	Collections []client.Collection
	// indicates if the node is Closed.
	Closed bool
}

// State contains all testing State.
type State struct {
	// The test context.
	Ctx context.Context

	// The Go Test test state
	T testing.TB

	// The type of KMS currently being tested.
	Kms KMSType

	// The type of database currently being tested.
	Dbt DatabaseType

	// The type of client currently being tested.
	ClientType ClientType

	// Any explicit transactions active in this test.
	//
	// This is order dependent and the property is accessed by index.
	Txns []client.Txn

	// IdentityTypes is a map of identity to key type.
	// Use it to customize the key type that is used for identity and signing.
	IdentityTypes map[Identity]crypto.KeyType

	// Identities contains all Identities created in this test.
	// The map key is the identity reference that uniquely identifies Identities of different
	// types. See [identRef].
	// The map value is the identity holder that contains the identity itself and token
	// generated for different target nodes. See [identityHolder].
	Identities map[Identity]*IdentityHolder

	// The seed for the next identity generation. We want identities to be deterministic.
	NextIdentityGenSeed int

	// Policy IDs, by node index, by policyID index (in the order they were added).
	//
	// Note: In case acp type is sourcehub, all nodes will have the same state of PolicyIDs.
	PolicyIDs [][]string

	// Will receive an item once all actions have finished processing.
	AllActionsDone chan struct{}

	// These channels will receive a function which asserts results of any subscription requests.
	SubscriptionResultsChans []chan func()

	// The Nodes active in this test.
	Nodes []*NodeState

	// The ACP options to share between each node.
	DocumentACPOptions []node.DocumentACPOpt

	// The names of the collections active in this test.
	// Indexes matches that of initial collections.
	CollectionNames []string

	// A map of the collection indexes by their CollectionID, this allows easier
	// identification of collections in a natural, human readable, order
	// even when they are renamed.
	CollectionIndexesByCollectionID map[string]int

	// Document IDs by index, by collection index.
	//
	// Each index is assumed to be global, and may be expected across multiple
	// nodes.
	DocIDs [][]client.DocID

	// IsBench indicates wether the test is currently being benchmarked.
	IsBench bool

	// The SourceHub address used to pay for SourceHub transactions.
	SourcehubAddress string

	// IsNetworkEnabled indicates whether the network is enabled.
	IsNetworkEnabled bool

	// StatefulMatchers contains all stateful matchers that have been executed during a single
	// test run. After a single test run, the StatefulMatchers are reset.
	StatefulMatchers []StatefulMatcher

	// node id that is currently being asserted. This is used by [StatefulMatcher]s to know for which
	// node they should be asserting. For example, the [UniqueValue] matcher checks that it is
	// called with a value that it didn't see before, but the value should be the same for different
	// nodes, e.g. within the same node Cids should be unique, but across different nodes the same block
	// should have the same Cid.
	CurrentNodeID int
}

func (s *State) GetClientType() ClientType {
	return s.ClientType
}

func (s *State) GetCurrentNodeID() int {
	return s.CurrentNodeID
}

func (s *State) GetIdentity(ident Identity) acpIdentity.Identity {
	return GetIdentity(s, immutable.Some(ident))
}

var _ TestState = &State{}

// NewState returns a new fresh state for the given testCase.
func NewState(
	ctx context.Context,
	t testing.TB,
	testCase TestCase,
	kms KMSType,
	dbt DatabaseType,
	clientType ClientType,
	collectionNames []string,
) *State {
	s := &State{
		Ctx:                             ctx,
		T:                               t,
		Kms:                             kms,
		Dbt:                             dbt,
		ClientType:                      clientType,
		Txns:                            []client.Txn{},
		IdentityTypes:                   testCase.IdentityTypes,
		Identities:                      map[Identity]*IdentityHolder{},
		NextIdentityGenSeed:             0,
		AllActionsDone:                  make(chan struct{}),
		SubscriptionResultsChans:        []chan func(){},
		Nodes:                           []*NodeState{},
		DocumentACPOptions:              []node.DocumentACPOpt{},
		CollectionNames:                 collectionNames,
		CollectionIndexesByCollectionID: map[string]int{},
		DocIDs:                          [][]client.DocID{},
		PolicyIDs:                       [][]string{},
		IsBench:                         false,
	}
	return s
}
