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

// Blocks carry the id of the collection version they were authored against. Once the collection is
// patched again that version is neither the root nor the active version, and the serving node must
// still resolve the collection to authorize the blocks.
func TestACP_P2PSubscribeUpdateAuthoredOnInactiveCollectionVersion_SyncsUpdate(t *testing.T) {
	test := testUtils.TestCase{
		// Releases before v1.2.0 resolve the collection by the block's version id.
		// https://github.com/sourcenetwork/defradb/issues/5303
		OldFirstNodeSupportedFromVersion: "v1.2.0",
		OldLastNodeSupportedFromVersion:  "v1.2.0",
		SupportedDocumentACPTypes: immutable.Some(
			[]state.DocumentACPType{
				state.LocalDocumentACPType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy: `
description: a test policy which marks a collection in a database as a resource
name: test
resources:
- name: users
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: reader
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
					}
				`,
			},
			&action.AddDoc{
				NodeID:   immutable.Some(0),
				Identity: testUtils.ClientIdentity(1),
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.AddDACActorRelationship{
				NodeID:            immutable.Some(0),
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NodeIdentity(1),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedExistence: false,
			},
			&action.PatchCollection{
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			&action.UpdateDoc{
				// This block is authored against the second collection version.
				NodeID:   immutable.Some(0),
				Identity: testUtils.ClientIdentity(1),
				Doc: `{
					"name": "Fred"
				}`,
			},
			&action.PatchCollection{
				// The second collection version is now neither the root nor the active version.
				NodeID: immutable.Some(0),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} }
					]
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			&action.UpdateDoc{
				// Syncing this head requires node 1 to fetch the blocks of the earlier update from node 0.
				NodeID:   immutable.Some(0),
				Identity: testUtils.ClientIdentity(1),
				Doc: `{
					"email": "fred@source.network"
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID:   immutable.Some(1),
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
