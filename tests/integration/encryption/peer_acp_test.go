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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

const policy = `
name: Test Policy

description: A Policy

actor:
  name: actor

resources:
  users:
    permissions:
      read:
        expr: owner + reader + updater + deleter

      update:
        expr: owner + updater

      delete:
        expr: owner + deleter

      nothing:
        expr: dummy

    relations:
      owner:
        types:
          - actor

      reader:
        types:
          - actor

      updater:
        types:
          - actor

      deleter:
        types:
          - actor

      admin:
        manages:
          - reader
        types:
          - actor

      dummy:
        types:
          - actor
`

func TestDocEncryptionACP_IfUserAndNodeHaveAccess_ShouldFetch(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		SupportedDocumentACPTypes: immutable.Some(
			[]testUtils.DocumentACPType{
				testUtils.SourceHubDocumentACPType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(0),
				Policy:   policy,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,

				Replace: map[string]testUtils.ReplaceType{
					"Policy0": testUtils.NewPolicyIndex(0),
				},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
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
			testUtils.WaitForSync{
				Decrypted: []int{0},
			},
			testUtils.Request{
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
			[]testUtils.DocumentACPType{
				testUtils.SourceHubDocumentACPType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(0),
				Policy:   policy,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,

				Replace: map[string]testUtils.ReplaceType{
					"Policy0": testUtils.NewPolicyIndex(0),
				},
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
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
			testUtils.Request{
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
			// If the node doesn't have rights to the doc, it can't do merge
			// and therefore has no heads. So commits should be empty.
			testUtils.Request{
				NodeID:   immutable.Some(1),
				Identity: testUtils.ClientIdentity(1),
				Request: `
					query {
						commits {
							delta
							docID
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{},
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
			[]testUtils.DocumentACPType{
				testUtils.SourceHubDocumentACPType,
			},
		),
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.AddDACPolicy{
				Identity: testUtils.NodeIdentity(0),
				Policy:   policy,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,

				Replace: map[string]testUtils.ReplaceType{
					"Policy0": testUtils.NewPolicyIndex(0),
				},
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
			testUtils.CreateDoc{
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
			testUtils.CreateDoc{
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
			testUtils.CreateDoc{
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
			testUtils.CreateDoc{
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
			testUtils.CreateDoc{
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
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `
					{
						"name": "Shahzad",
						"age": 33
					}
				`,
			},
			testUtils.WaitForSync{
				Decrypted: []int{0, 2},
			},
			testUtils.Request{
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
						{"name": "John"},
						{"name": "Islam"},
						{"name": "Shahzad"},
						{"name": "Fred"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionACP_IfClientNodeHasDocPermissionButServerNodeIsNotAvailable_ShouldNotFetch(t *testing.T) {
	test := testUtils.TestCase{
		KMS: testUtils.KMS{Activated: true},
		SupportedDocumentACPTypes: immutable.Some(
			[]testUtils.DocumentACPType{
				testUtils.SourceHubDocumentACPType,
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
			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,

				Replace: map[string]testUtils.ReplaceType{
					"Policy0": testUtils.NewPolicyIndex(0),
				},
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
			testUtils.CreateDoc{
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
			testUtils.Request{
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
