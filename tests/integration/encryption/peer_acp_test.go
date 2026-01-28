// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encryption

import (
	"testing"
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

const policy = `
name: Test Policy

description: A Policy

resources:
  - name: users
    permissions:
    - name: read
      expr: reader + updater + deleter
    - name: update
      expr: updater
    - name: delete
      expr: deleter
    - name: nothing
      expr: dummy

    relations:
    - name: reader
      types:
      - actor
    - name: updater
      types:
      - actor
    - name: deleter
      types:
      - actor
    - name: admin
      manages:
      - reader
      types:
      - actor
    - name: dummy
      types:
      - actor
`

func TestDocEncryptionACP_IfUserAndNodeHaveAccess_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		SupportedDocumentACPTypes: immutable.Some(
			[]state.DocumentACPType{
				state.SourceHubDocumentACPType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(0),
				Policy:   policy,
			},
			&action.AddSchema{
				Schema: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			&action.CreateDoc{
				NodeID:   immutable.Some(0),
				Identity: testUtils.ClientIdentity(0),
				Doc: `
					{
						"name": "Fred",
						"age": 33
					}
				`,
				IsDocEncrypted: true,
			},
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(0),
				TargetIdentity:    testUtils.ClientIdentity(1),
				DocID:             0,
				Relation:          "reader",
			},
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(0),
				TargetIdentity:    testUtils.NodeIdentity(1),
				DocID:             0,
				Relation:          "reader",
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
						{"name": "Fred"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionACP_IfUserHasAccessButNotNode_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		SupportedDocumentACPTypes: immutable.Some(
			[]state.DocumentACPType{
				state.SourceHubDocumentACPType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(0),
				Policy:   policy,
			},
			&action.AddSchema{
				Schema: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			&action.CreateDoc{
				NodeID:   immutable.Some(0),
				Identity: testUtils.ClientIdentity(0),
				Doc: `
					{
						"name": "Fred",
						"age": 33
					}
				`,
				IsDocEncrypted: true,
			},
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(0),
				TargetIdentity:    testUtils.ClientIdentity(1),
				DocID:             0,
				Relation:          "reader",
			},
			testUtils.Wait{Duration: 100 * time.Millisecond},
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
					"Users": []map[string]any{},
				},
			},
			// If the instance doesn't have rights to the doc, it can't do block sync
			// and therefore doesn't have the related commit blocks.
			&action.Request{
				NodeID:   immutable.Some(1),
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						_commits {
							delta
							docID
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionACP_IfNodeHasAccessToSomeDocs_ShouldFetchOnlyThem(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		SupportedDocumentACPTypes: immutable.Some(
			[]state.DocumentACPType{
				state.SourceHubDocumentACPType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.AddDACPolicy{
				Identity: testUtils.NodeIdentity(0),
				Policy:   policy,
			},
			&action.AddSchema{
				Schema: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			// encrypted, private, shared
			&action.CreateDoc{
				NodeID:   immutable.Some(0),
				Identity: testUtils.NodeIdentity(0),
				Doc: `
					{
						"name": "Fred",
						"age": 33
					}
				`,
				IsDocEncrypted: true,
			},
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.NodeIdentity(0),
				TargetIdentity:    testUtils.NodeIdentity(1),
				DocID:             0,
				Relation:          "reader",
			},
			// encrypted, private, not shared
			&action.CreateDoc{
				NodeID:   immutable.Some(0),
				Identity: testUtils.NodeIdentity(0),
				Doc: `
					{
						"name": "Andy",
						"age": 33
					}
				`,
				IsDocEncrypted: true,
			},
			// encrypted, public
			&action.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `
					{
						"name": "Islam",
						"age": 33
					}
				`,
				IsDocEncrypted: true,
			},
			// not encrypted, private, shared
			&action.CreateDoc{
				NodeID:   immutable.Some(0),
				Identity: testUtils.NodeIdentity(0),
				Doc: `
					{
						"name": "John",
						"age": 33
					}
				`,
			},
			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.NodeIdentity(0),
				TargetIdentity:    testUtils.NodeIdentity(1),
				DocID:             3,
				Relation:          "reader",
			},
			// not encrypted, private, not shared
			&action.CreateDoc{
				NodeID:   immutable.Some(0),
				Identity: testUtils.NodeIdentity(0),
				Doc: `
					{
						"name": "Keenan",
						"age": 33
					}
				`,
			},
			// not encrypted, public
			&action.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `
					{
						"name": "Shahzad",
						"age": 33
					}
				`,
			},
			testUtils.WaitForSync{},
			&action.Request{
				NodeID:   immutable.Some(1),
				Identity: testUtils.NodeIdentity(1),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Fred"},
						{"name": "John"},
						{"name": "Islam"},
						{"name": "Shahzad"},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionACP_IfClientNodeHasDocPermissionButServerNodeIsNotAvailable_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		SupportedDocumentACPTypes: immutable.Some(
			[]state.DocumentACPType{
				state.SourceHubDocumentACPType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.AddDACPolicy{
				Identity: testUtils.NodeIdentity(0),
				Policy:   policy,
			},
			&action.AddSchema{
				Schema: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 2,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        2,
				CollectionIDs: []int{0},
			},
			&action.CreateDoc{
				NodeID:   immutable.Some(0),
				Identity: testUtils.NodeIdentity(0),
				Doc: `
					{
						"name": "Fred",
						"age": 33
					}
				`,
				IsDocEncrypted: true,
			},
			testUtils.Close{
				NodeID: immutable.Some(0),
			},
			testUtils.AddDACActorRelationship{
				NodeID:            immutable.Some(1),
				RequestorIdentity: testUtils.NodeIdentity(0),
				TargetIdentity:    testUtils.NodeIdentity(1),
				DocID:             0,
				Relation:          "reader",
			},
			testUtils.Wait{
				Duration: 100 * time.Millisecond,
			},
			&action.Request{
				NodeID:   immutable.Some(1),
				Identity: testUtils.NodeIdentity(1),
				Request: `
					query {
						Users {
							name
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
