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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/net"
)

type state struct {
	// The test context.
	ctx context.Context

	// The Go Test test state
	t *testing.T

	// The TestCase currently being executed.
	testCase TestCase

	// The type of database currently being tested.
	dbt DatabaseType

	// Any explicit transactions active in this test.
	//
	// This is order dependent and the property is accessed by index.
	txns []datastore.Txn

	// Will recieve an item once all actions have finished processing.
	allActionsDone chan struct{}

	// These channels will recieve a function which asserts results of any subscription requests.
	subscriptionResultsChans []chan func()

	// These synchronisation channels allow async actions to track their completion.
	syncChans []chan struct{}

	// The addresses of any nodes configured.
	nodeAddresses []string

	// The configurations for any nodes
	nodeConfigs []config.Config

	// The nodes active in this test.
	nodes []*net.Node

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
}

// newState returns a new fresh state for the given testCase.
func newState(
	ctx context.Context,
	t *testing.T,
	testCase TestCase,
	dbt DatabaseType,
	collectionNames []string,
) *state {
	return &state{
		ctx:                      ctx,
		t:                        t,
		testCase:                 testCase,
		dbt:                      dbt,
		txns:                     []datastore.Txn{},
		allActionsDone:           make(chan struct{}),
		subscriptionResultsChans: []chan func(){},
		syncChans:                []chan struct{}{},
		nodeAddresses:            []string{},
		nodeConfigs:              []config.Config{},
		nodes:                    []*net.Node{},
		dbPaths:                  []string{},
		collections:              [][]client.Collection{},
		collectionNames:          collectionNames,
		documents:                [][]*client.Document{},
	}
}
