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

// The policy used by both tests below. Grants "updater" both update and
// read (via "read: reader + updater + deleter") so a non-creator with the
// "updater" relation can both update and read the document.
const policyForUpdaterTests = `
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
    - updater
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
`

// A non-creator on a peer node updates an indexed field on a doc that
// arrived via P2P. Before the fix this hit "corrupted index" because the
// merge processor fetched the post-merge doc through the ACP read filter
// with no caller identity, got nil, and silently skipped the index Save.
// See issue #4834.
func TestACP_P2P_NonCreatorUpdatesIndexedFieldOnPeerNode_Succeeds(t *testing.T) {
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
				Policy:   policyForUpdaterTests,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String @index
						age: Int
					}
				`,
			},

			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},

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

			testUtils.AddDACActorRelationship{
				NodeID:            immutable.Some(0),
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				CollectionID:      0,
				DocID:             0,
				Relation:          "updater",
				ExpectedExistence: false,
			},

			// Node 1 needs the same grant to accept the inbound commit (DGC-012).
			testUtils.AddDACActorRelationship{
				NodeID:            immutable.Some(0),
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NodeIdentity(1),
				CollectionID:      0,
				DocID:             0,
				Relation:          "updater",
				ExpectedExistence: false,
			},

			testUtils.WaitForSync{},

			&action.UpdateDoc{
				Identity:     testUtils.ClientIdentity(2),
				NodeID:       immutable.Some(1),
				CollectionID: 0,
				DocID:        0,
				Doc: `
					{
						"name": "Alice Updated"
					}
				`,
			},

			// Read back as the owner to assert data + index integrity.
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
						{"name": "Alice Updated", "age": int64(30)},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Control for the test above: same scenario without @index. Confirms
// the "non-creator updates on peer node" path itself was fine; the
// indexed-vs-not difference is what the indexed test is exercising.
// Also closes a coverage gap — no existing P2P DAC test had a
// non-creator UpdateDoc on the receiving node.
func TestACP_P2P_NonCreatorUpdatesNonIndexedFieldOnPeerNode_Succeeds(t *testing.T) {
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
				Policy:   policyForUpdaterTests,
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

			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},

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

			testUtils.AddDACActorRelationship{
				NodeID:            immutable.Some(0),
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				CollectionID:      0,
				DocID:             0,
				Relation:          "updater",
				ExpectedExistence: false,
			},

			testUtils.AddDACActorRelationship{
				NodeID:            immutable.Some(0),
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NodeIdentity(1),
				CollectionID:      0,
				DocID:             0,
				Relation:          "updater",
				ExpectedExistence: false,
			},

			testUtils.WaitForSync{},

			&action.UpdateDoc{
				Identity:     testUtils.ClientIdentity(2),
				NodeID:       immutable.Some(1),
				CollectionID: 0,
				DocID:        0,
				Doc: `
					{
						"age": 31
					}
				`,
			},

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
						{"name": "Alice", "age": int64(31)},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
