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

func TestACP_P2PUpdatePrivateDocumentsOnDifferentNodes_ConflictingPermission_SourceHubACP(t *testing.T) {
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
name: Test Policy
resources:
- name: users
  permissions:
  - name: update
    expr: updater
  - name: read
    expr: updater + reader
  - name: delete
  relations:
  - name: updater
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
							age: Int
						}
					`,
			},
			&action.AddDoc{
				Identity:     testUtils.ClientIdentity(1),
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "Mister Andy",
				},
			},
			// authorize both nodes to receive document
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NodeIdentity(1),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedExistence: false,
			},
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.NodeIdentity(0),
				CollectionID:      0,
				DocID:             0,
				Relation:          "reader",
				ExpectedExistence: false,
			},
			// authorize colalborator to update document
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),
				TargetIdentity:    testUtils.ClientIdentity(2),
				CollectionID:      0,
				DocID:             0,
				Relation:          "updater",
				ExpectedExistence: false,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.SyncDocs{
				NodeID:       1,
				CollectionID: 0,
				DocIDs:       []int{0},
				SourceNodes:  []int{0},
			},
			testUtils.WaitForSync{},
			// all nodes should have the original doc
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Mister Andy"},
					},
				},
			},
			// peer will update the doc while they have permission
			testUtils.UpdateDoc{
				Identity:     testUtils.ClientIdentity(2),
				NodeID:       immutable.Some(1),
				CollectionID: 0,
				DocID:        0,
				Doc: `
						{
							"name": "Sir Andy"
						}
					`,
			},
			// remove update permission from peer
			testUtils.DeleteDACActorRelationship{
				RequestorIdentity:   testUtils.ClientIdentity(1),
				TargetIdentity:      testUtils.ClientIdentity(2),
				CollectionID:        0,
				DocID:               0,
				Relation:            "updater",
				ExpectedRecordFound: true,
			},
			// original node should receive the updated document, after the collaborator got his permission removed
			testUtils.SyncDocs{
				NodeID:       0,
				CollectionID: 0,
				DocIDs:       []int{0},
				SourceNodes:  []int{1},
			},
			testUtils.WaitForSync{},
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				NodeID:   immutable.Some(0),
				Request: `query {
						Users {
							name
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Sir Andy"},
					},
				},
			},
			&action.Request{
				Identity: testUtils.ClientIdentity(1),
				NodeID:   immutable.Some(1),
				Request: `query {
						Users {
							name
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Sir Andy"},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
