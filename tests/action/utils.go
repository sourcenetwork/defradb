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
	"strings"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
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

// refreshCollections refreshes all the collections of the given names, preserving order.
//
// If a given collection is not present in the database the value at the corresponding
// result-index will be nil.
func refreshCollections(
	s *state.State,
) {
	nodeIDs, nodes := getNodesWithIDs(immutable.None[int](), s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		// Inject node's identity into the context while refreshing so the [GetCollections] call
		// doesn't fail due to lack of authorization(s) if NAC is enabled.
		nodeIdentity := NodeIdentity(nodeID)
		node.Collections = make([]client.Collection, len(s.CollectionNames))
		ctx := getContextWithIdentity(s.Ctx, s, nodeIdentity, nodeID)
		allCollections, err := node.GetCollections(ctx, client.CollectionFetchOptions{})
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
