// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package action

import (
	"context"
	"time"
)

// SyncBranchableCollection is an action that syncs a branchable collection's DAG
// from another node.
type SyncBranchableCollection struct {
	stateful

	// NodeID holds the ID (index) of a node to request the sync from.
	//
	// The collection will be synced to the other nodes that are connected
	// and subscribed to the collection's pubsub topics.
	NodeID int

	// CollectionName is the name of the collection to sync.
	CollectionName string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*SyncBranchableCollection)(nil)
var _ Stateful = (*SyncBranchableCollection)(nil)

func (a *SyncBranchableCollection) Execute() {
	ctx, cancel := context.WithTimeout(a.s.Ctx, 5*time.Second)
	defer cancel()

	node := a.s.Nodes[a.NodeID]
	err := node.SyncBranchableCollection(ctx, a.CollectionName)

	expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
	assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
}
