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

package peer

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// TestP2PLateSub_BookWithRelation_RelationSkippedOnMerge_NoError asserts that when
// a node subscribes to a document whose relation target it has never received,
// the P2P merge succeeds. validateMergeRelationDocIDs treats a missing target as
// a skip (the target may arrive later), so the merge is not rejected.
//
// Scenario:
//  1. Node A creates Author and Book (linked). Node B does not subscribe yet.
//  2. Peers connect. Node B subscribes only to the Book document.
//  3. Node A updates the Book, triggering a sync event to Node B.
//  4. Node B receives the merge but does not have Author locally → dangling link.
//  5. The merge must succeed: Node B stores the Book with its _AuthorID intact
//     even though Author is absent.
func TestP2PLateSub_BookWithRelation_RelationSkippedOnMerge_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Author {
						Name: String
						Books: [Book]
					}
					type Book {
						Name: String
						Author: Author
					}
				`,
			},
			// Node 0: create Author (not synced because NodePeers don't auto-sync new docs).
			&action.AddDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				Doc:          `{"Name": "Frank Herbert"}`,
			},
			// All nodes: create Book — both nodes will have this document.
			&action.AddDoc{
				CollectionID: 1,
				Doc:          `{"Name": "Dune"}`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			// Node 1 subscribes to the Book document so it will receive future updates.
			// Author was never synced to Node 1.
			testUtils.AddDocumentSubscription{
				NodeID: 1,
				DocIDs: []state.ColDocIndex{
					state.NewColDocIndex(1, 0),
				},
			},
			// Node 0 links the Book to the Author. This triggers a sync event to Node 1.
			// DocMap uses NewDocIndex so the Author's real DocID is substituted at runtime.
			&action.UpdateDoc{
				NodeID:       immutable.Some(0),
				CollectionID: 1,
				DocID:        0,
				DocMap: map[string]any{
					"_AuthorID": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.WaitForSync{},
			// Node 1 should have received and stored the Book update.
			// Author is absent, but the merge must have succeeded (dangling link is
			// acceptable on the P2P merge path).
			&action.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Book {
						Name
						Author {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Book": []map[string]any{
						{
							"Name": "Dune",
							// Author not available on Node 1 — merge succeeded anyway.
							"Author": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
