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

// AddSchema is an action that will add the given GQL schema to the Defra nodes.
type SyncCollection struct {
	stateful

	// NodeID holds the ID (index) of a node to sync collections to.
	NodeID int

	// VersionIDs to pass into the `SyncCollections` call.
	VersionIDs []string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*SyncCollection)(nil)
var _ Stateful = (*SyncCollection)(nil)

func (a *SyncCollection) Execute() {
	replacedVersionIDs := replaceMap(a.s, 0, a.VersionIDs)
	versionIDs := make([]string, len(a.VersionIDs))
	for i, originalID := range a.VersionIDs {
		versionIDs[i] = replacedVersionIDs[originalID]
	}

	ctx, cancel := context.WithTimeout(a.s.Ctx, 5*time.Second)
	defer cancel()

	node := a.s.Nodes[a.NodeID]
	err := node.SyncCollections(ctx, versionIDs...)

	expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
	assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)

	// If the schema was updated we need to refresh the collection definitions.
	refreshCollections(a.s)
}
