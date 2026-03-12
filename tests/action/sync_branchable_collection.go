// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package action

import (
	"context"
	"errors"
	"time"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
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

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// CollectionID is the index of the collection to sync.
	CollectionID int

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*SyncBranchableCollection)(nil)
var _ Stateful = (*SyncBranchableCollection)(nil)

func (a *SyncBranchableCollection) Execute() {
	ctx, cancel := context.WithTimeout(a.s.Ctx, time.Second)
	defer cancel()

	nodeState := a.s.Nodes[a.NodeID]

	if a.CollectionID >= len(nodeState.Collections) {
		err := assertError(a.s.T,
			errors.New("collection index out of range"),
			a.ExpectedError)
		assertExpectedErrorRaised(a.s.T, a.ExpectedError, err)
		return
	}

	opts := options.SyncBranchableCollection()
	identOption := getIdentityForRequestSpecificToNode(a.s, a.Identity, a.NodeID)
	if identOption.HasValue() {
		opts.SetIdentity(identOption.Value())
	}

	collection := nodeState.Collections[a.CollectionID]
	err := nodeState.SyncBranchableCollection(ctx, collection.CollectionID(), opts)

	expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
	assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)
}
