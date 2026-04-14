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
	"strings"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/state"
)

func getNodesWithIDs(nodeID immutable.Option[int], nodes []*state.NodeState) ([]int, []*state.NodeState) {
	if !nodeID.HasValue() {
		indexes := make([]int, len(nodes))
		for i := range nodes {
			indexes[i] = i
		}
		return indexes, nodes
	}

	return []int{nodeID.Value()}, []*state.NodeState{nodes[nodeID.Value()]}
}

// RefreshCollections refreshes all the collections of the given names, preserving order.
// If a given collection is not present in the database the value at the corresponding
// result-index will be nil.
func RefreshCollections(
	s *state.State,
) {
	nodeIDs, nodes := getNodesWithIDs(immutable.None[int](), s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		// Create a txn for this refresh
		txn, err := node.NewTxn(false)
		defer txn.Discard()
		require.Nil(s.T, err)
		ctx := db.InitContext(s.Ctx, txn)

		// Inject node's identity into the context and options while refreshing so the [GetCollections] call
		// doesn't fail due to lack of authorization(s) if NAC is enabled.
		nodeIdentity := NodeIdentity(nodeID)
		node.Collections = make([]client.Collection, len(s.CollectionNames))
		identOption := getIdentityForRequestSpecificToNode(s, nodeIdentity, nodeID)
		opts := options.GetCollections()
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}
		allCollections, err := txn.GetCollections(ctx, opts)
		require.Nil(s.T, err)

		for i, collectionName := range s.CollectionNames {
			for _, collection := range allCollections {
				if collection.Name() == collectionName {
					if _, ok := s.CollectionIndexesByCollectionID[collection.Version().CollectionID]; !ok {
						// If the root is not found here this is likely the first refreshCollections
						// call of the test, we map it by root in case the collection is renamed -
						// we still wish to preserve the original index so test maintainers can reference
						// them in a convenient manner.
						s.CollectionIndexesByCollectionID[collection.Version().CollectionID] = i
					}
					break
				}
			}
		}

		for _, collection := range allCollections {
			if index, ok := s.CollectionIndexesByCollectionID[collection.Version().CollectionID]; ok {
				node.Collections[index] = collection
			}
		}
	}
}

// GetCanonicallyOrderedCollections gets the collections inside of a transaction, if one is provided.
// If one is not provided, it will default to running the GetCollections function on the node itself.
// Importantly, this will use the same ordering as would be found in the node.Collections slice that
// is refreshed by the RefreshCollections function.
func GetCanonicallyOrderedCollections(
	s *state.State,
	node *state.NodeState,
	txn immutable.Option[client.Txn],
) []client.Collection {
	var clientTxn client.Txn
	if txn.HasValue() {
		clientTxn = txn.Value()
	}

	// Find the nodeID for this node
	nodeID := -1
	for i, n := range s.Nodes {
		if n == node {
			nodeID = i
			break
		}
	}

	nodeIdentity := NodeIdentity(nodeID)

	newCollections := make([]client.Collection, len(s.CollectionNames))

	identOption := getIdentityForRequestSpecificToNode(s, nodeIdentity, nodeID)
	opts := options.GetCollections()
	if identOption.HasValue() {
		opts.SetIdentity(identOption.Value())
	}

	var allCollections []client.Collection
	var err error

	if clientTxn != nil {
		allCollections, err = clientTxn.GetCollections(s.Ctx, opts)
	} else {
		allCollections, err = node.GetCollections(s.Ctx, opts)
	}
	require.Nil(s.T, err)

	for i, collectionName := range s.CollectionNames {
		for _, collection := range allCollections {
			if collection.Name() == collectionName {
				if _, ok := s.CollectionIndexesByCollectionID[collection.Version().CollectionID]; !ok {
					s.CollectionIndexesByCollectionID[collection.Version().CollectionID] = i
				}
				break
			}
		}
	}

	for _, collection := range allCollections {
		if index, ok := s.CollectionIndexesByCollectionID[collection.Version().CollectionID]; ok {
			newCollections[index] = collection
		}
	}

	return newCollections
}

func appendCollectionVersion(s *state.State, versionID string) {
	for _, existingVersion := range s.CollectionVersions {
		if existingVersion == versionID {
			return
		}
	}

	s.CollectionVersions = append(s.CollectionVersions, versionID)
}

// withRetryOnNode attempts to perform the given action, retrying up to a DB-defined
// maximum attempt count if a transaction conflict error is returned.
//
// If a P2P-sync commit for the given document is already in progress this
// Save call can fail as the transaction will conflict. We dont want to worry
// about this in our tests so we just retry a few times until it works (or the
// retry limit is breached - important incase this is a different error)
func withRetryOnNode(
	node clients.Client,
	action func() error,
) error {
	for i := 0; i < node.MaxTxnRetries(); i++ {
		err := action()
		// Check the contents of the error instead of the type, because it may have
		// lost its type while passing through the C binding layer.
		if err != nil && strings.Contains(err.Error(), corekv.ErrTxnConflict.Error()) {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		return err
	}
	return nil
}
