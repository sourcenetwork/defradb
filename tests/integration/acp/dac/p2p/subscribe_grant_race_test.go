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

package test_acp_dac_p2p

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

// This mirrors a downstream group-chat timeline: on the subscription
// (pull) path, a policy-gated document is created on the source node BEFORE the read grant for
// the eventual non-creator reader and the grant for the receiving node have propagated. The
// receiver's pushlog handler then walks the document's DAG, and the source serves each linked
// block through a per-block access check. With several updates the document spans multiple
// blocks, so the source runs the access check repeatedly for the same (peer, document) pair.
//
// The test asserts the user-visible outcome the race breaks: once the grants land, the
// non-creator can read the fully-synced document on the receiving node.
func TestACP_P2PSubscribe_MultiBlockDoc_NonCreatorReadsAfterGrantsLand_SourceHubACP(t *testing.T) {
	test := testUtils.TestCase{
		SupportedDocumentACPTypes: immutable.Some(
			[]state.DocumentACPType{
				state.SourceHubDocumentACPType,
			},
		),

		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),

			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy: `
description: A Policy
name: Test Policy
resources:
- name: users
  permissions:
  - expr: deleter
    name: delete
  - expr: dummy
    name: nothing
  - expr: reader + updater + deleter
    name: read
  - expr: updater
    name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: deleter
    types:
    - actor
  - name: dummy
    types:
    - actor
  - name: reader
    types:
    - actor
  - name: updater
    types:
    - actor
`,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			// Pull-based: the receiver (node 1) subscribes to the collection on node 0.
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},

			// Document created before any non-creator/receiver grants exist.
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				Doc: `
					{
						"name": "Alice",
						"age": 30
					}
				`,
			},

			// Several updates so the document's DAG spans multiple blocks; serving them all is
			// what previously incurred one access-control round-trip per block on the source.
			&action.UpdateDoc{
				Identity:     testUtils.ClientIdentity(1),
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				DocID:        0,
				Doc:          `{ "age": 31 }`,
			},
			&action.UpdateDoc{
				Identity:     testUtils.ClientIdentity(1),
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				DocID:        0,
				Doc:          `{ "age": 32 }`,
			},
			&action.UpdateDoc{
				Identity:     testUtils.ClientIdentity(1),
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				DocID:        0,
				Doc:          `{ "age": 33 }`,
			},

			// Grant the non-creator reader.
			testUtils.AddDACActorRelationship{
				NodeID:            immutable.Some(0),
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedExistence: false,
			},

			// Grant the receiving node so it may pull the blocks.
			testUtils.AddDACActorRelationship{
				NodeID:            immutable.Some(0),
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NodeIdentity(1),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedExistence: false,
			},

			testUtils.WaitForSync{},

			// Owner can read the fully-updated document on the receiver, confirming the whole DAG
			// replicated, not just the genesis block.
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				NodeID:   immutable.Some(1),
				Request: `
					query {
						Users {
							name
							age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Alice", "age": int64(33)},
					},
				},
			},

			// The key assertion: the non-creator with a reader grant sees the document on the
			// receiving node.
			&action.Request{
				Identity: testUtils.ClientIdentity(2),
				NodeID:   immutable.Some(1),
				Request: `
					query {
						Users {
							name
							age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Alice", "age": int64(33)},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
