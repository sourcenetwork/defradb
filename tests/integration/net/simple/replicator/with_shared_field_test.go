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

package replicator

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// Two documents with the same field value share one genesis field block (its CID is owned by
// both documents). When those documents are synced to another node, the merge must attribute
// the shared field block to each document, so the shared value is returned for both - not just
// for one of them.
func TestP2PReplicator_SharedFieldValue_SyncsForBothDocuments(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			// Same Name (shared "Name" field block), different Age (unique "Age" field block).
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Shared",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"Name": "Shared",
					"Age": 22
				}`,
			},
			testUtils.WaitForSync{},
			// On both the source and the replicated node, each document exposes the shared
			// "Name" value alongside its own "Age".
			&action.Request{
				Request: `query {
					Users {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"Name": "Shared", "Age": int64(21)},
						{"Name": "Shared", "Age": int64(22)},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
