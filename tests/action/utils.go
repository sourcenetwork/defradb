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
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
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
	for _, node := range s.Nodes {
		node.Collections = make([]client.Collection, len(s.CollectionNames))
		allCollections, err := node.GetCollections(s.Ctx, client.CollectionFetchOptions{})
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
