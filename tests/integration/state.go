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

	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/datastore"
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
}

// newState returns a new fresh state for the given testCase.
func newState(
	ctx context.Context,
	t *testing.T,
	testCase TestCase,
	dbt DatabaseType,
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
	}
}
