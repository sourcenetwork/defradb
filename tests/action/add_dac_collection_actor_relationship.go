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
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/state"
)

// AddDACCollectionActorRelationship will attempt to add a new relationship between a branchable
// collection's commit DAG (the collection object) and an actor.
//
// This is the collection-level analogue of [AddDACActorRelationship]: instead of sharing access to
// a single document, it shares access to a branchable collection's collection-level commit DAG.
type AddDACCollectionActorRelationship struct {
	stateful

	// NodeID may hold the ID (index) of the node we want to add the collection actor relationship on.
	//
	// If a value is not provided the relationship will be added in all nodes, unless testing with
	// sourcehub ACP, in which case the relationship will only be defined once.
	NodeID immutable.Option[int]

	// The collection whose commit DAG we want to add a relationship for.
	//
	// This is a required field. To test the invalid usage of not having this arg, use -1 index.
	CollectionID int

	// The name of the relation to set between collection and target actor (should be defined in the policy).
	//
	// This is a required field.
	Relation string

	// The target public identity, i.e. the identity of the actor to tie the collection's relation with.
	//
	// This is a required field.
	TargetIdentity immutable.Option[state.Identity]

	// The requestor identity, i.e. identity of the actor adding the relationship.
	// Note: This identity must either own or have managing access defined in the policy.
	//
	// This is a required field.
	RequestorIdentity immutable.Option[state.Identity]

	// Result returns true if it was a no-op due to existing before, and false if a new relationship was made.
	ExpectedExistence bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

var _ Action = (*AddDACCollectionActorRelationship)(nil)
var _ Stateful = (*AddDACCollectionActorRelationship)(nil)

func (a *AddDACCollectionActorRelationship) Execute() {
	nodeIDs, nodes := getNodesWithIDs(a.NodeID, a.s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		collectionName, collectionObjectID := getCollectionAndCollectionObjectID(a.s, a.CollectionID, nodeID)

		opt := options.WithIdentity(options.AddDACActorRelationship(),
			getIdentityForRequestSpecificToNode(a.s, a.RequestorIdentity, nodeID))
		// The collection-level commit DAG is gated using the collection id as the acp object id,
		// so we pass it here in place of a docID.
		exists, err := node.AddDACActorRelationship(
			a.s.Ctx,
			collectionName,
			collectionObjectID,
			a.Relation,
			state.GetIdentityDID(a.s, a.TargetIdentity),
			opt,
		)

		expectedErrorRaised := assertError(a.s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(a.s.T, a.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(a.s.T, a.ExpectedError, "")
			require.Equal(a.s.T, a.ExpectedExistence, exists.ExistedAlready)
		}

		// The relationship should only be added to a SourceHub chain once - there is no need to loop
		// through the nodes.
		if a.s.DocumentACPType == state.SourceHubDocumentACPType {
			break
		}
	}
}

// getCollectionAndCollectionObjectID returns the collection name and the acp object id used to gate
// the collection-level commit DAG (the collection's CollectionID).
func getCollectionAndCollectionObjectID(s *state.State, collectionID, nodeID int) (string, string) {
	if collectionID == -1 {
		return "", ""
	}
	collection := s.Nodes[nodeID].Collections[collectionID]
	if collection.Version().Name == "" {
		require.Fail(s.T, "Expected non-empty collection name, but it was empty.")
	}
	// CollectionID is the acp object id used for collection-level checks; a "" would silently
	// validate the wrong scope.
	if collection.Version().CollectionID == "" {
		require.Fail(s.T, "Expected non-empty collection object ID, but it was empty.")
	}
	return collection.Version().Name, collection.Version().CollectionID
}
