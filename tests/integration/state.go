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

	identity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/event"
	"github.com/sourcenetwork/defradb/net"
	"github.com/sourcenetwork/defradb/tests/clients"
)

type state struct {
	// The test context.
	ctx context.Context

	// The Go Test test state
	t testing.TB

	// The TestCase currently being executed.
	testCase TestCase

	// The type of database currently being tested.
	dbt DatabaseType

	// The type of client currently being tested.
	clientType ClientType

	// Any explicit transactions active in this test.
	//
	// This is order dependent and the property is accessed by index.
	txns []datastore.Txn

	identities []identity.Identity

	// Will recieve an item once all actions have finished processing.
	allActionsDone chan struct{}

	// These channels will recieve a function which asserts results of any subscription requests.
	subscriptionResultsChans []chan func()

	// nodeMergeCompleteSubs is a list of all merge complete event subscriptions
	nodeMergeCompleteSubs []*event.Subscription

	// nodeUpdateSubs is a list of all update event subscriptions
	nodeUpdateSubs []*event.Subscription

	// nodeConnections contains all connected nodes.
	//
	// The index of the slice is the node id. The map key is the connected node id.
	nodeConnections []map[int]struct{}

	// nodeReplicatorSources contains all active replicators.
	//
	// The index of the slice is the source node id. The map key is the target node id.
	nodeReplicatorSources []map[int]struct{}

	// nodeReplicatorTargets contains all active replicators.
	//
	// The index of the slice is the target node id. The map key is the source node id.
	nodeReplicatorTargets []map[int]struct{}

	// nodePeerCollections contains all active peer collection subscriptions.
	//
	// The index of the slice is the collection id. The map key is the node id of the subscriber.
	nodePeerCollections []map[int]struct{}

	// actualDocHeads contains all document heads that exist on a node.
	//
	// The index of the slice is the node id. The map key is the doc id. The map value is the doc head.
	actualDocHeads []map[string]cid.Cid

	// expectedDocHeads contains all document heads that are expected to exist on a node.
	//
	// The index of the slice is the node id. The map key is the doc id. The map value is the doc head.
	expectedDocHeads []map[string]cid.Cid

	// The addresses of any nodes configured.
	nodeAddresses []peer.AddrInfo

	// The configurations for any nodes
	nodeConfigs [][]net.NodeOpt

	// The nodes active in this test.
	nodes []clients.Client

	// The paths to any file-based databases active in this test.
	dbPaths []string

	// Collections by index, by nodeID present in the test.
	// Indexes matches that of collectionNames.
	collections [][]client.Collection

	// The names of the collections active in this test.
	// Indexes matches that of collections.
	collectionNames []string

	// Documents by index, by collection index.
	//
	// Each index is assumed to be global, and may be expected across multiple
	// nodes.
	documents [][]*client.Document

	// Indexes, by index, by collection index, by node index.
	indexes [][][]client.IndexDescription

	// isBench indicates wether the test is currently being benchmarked.
	isBench bool
}

// newState returns a new fresh state for the given testCase.
func newState(
	ctx context.Context,
	t testing.TB,
	testCase TestCase,
	dbt DatabaseType,
	clientType ClientType,
	collectionNames []string,
) *state {
	return &state{
		ctx:                      ctx,
		t:                        t,
		testCase:                 testCase,
		dbt:                      dbt,
		clientType:               clientType,
		txns:                     []datastore.Txn{},
		allActionsDone:           make(chan struct{}),
		subscriptionResultsChans: []chan func(){},
		nodeMergeCompleteSubs:    []*event.Subscription{},
		nodeConnections:          []map[int]struct{}{},
		nodeReplicatorSources:    []map[int]struct{}{},
		nodeReplicatorTargets:    []map[int]struct{}{},
		nodePeerCollections:      []map[int]struct{}{},
		actualDocHeads:           []map[string]cid.Cid{},
		expectedDocHeads:         []map[string]cid.Cid{},
		nodeAddresses:            []peer.AddrInfo{},
		nodeConfigs:              [][]net.NodeOpt{},
		nodes:                    []clients.Client{},
		dbPaths:                  []string{},
		collections:              [][]client.Collection{},
		collectionNames:          collectionNames,
		documents:                [][]*client.Document{},
		indexes:                  [][][]client.IndexDescription{},
		isBench:                  false,
	}
}
