// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package signature

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/crypto"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocSignature_WithPeersAndSecp256k1KeyType_ShouldSync(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		IdentityTypes: map[testUtils.Identity]crypto.KeyType{
			testUtils.NodeIdentity(0).Value(): crypto.KeyTypeSecp256k1,
			testUtils.NodeIdentity(1).Value(): crypto.KeyTypeSecp256k1,
		},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type User {
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
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocSignature_WithPeersAndEd25519KeyType_ShouldSync(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		IdentityTypes: map[testUtils.Identity]crypto.KeyType{
			testUtils.NodeIdentity(0).Value(): crypto.KeyTypeEd25519,
			testUtils.NodeIdentity(1).Value(): crypto.KeyTypeEd25519,
		},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type User {
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
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					User {
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocSignature_WithPeersAnDifferentKeyTypes_ShouldSync(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		IdentityTypes: map[testUtils.Identity]crypto.KeyType{
			testUtils.NodeIdentity(0).Value(): crypto.KeyTypeSecp256k1,
			testUtils.NodeIdentity(1).Value(): crypto.KeyTypeEd25519,
		},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type User {
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
				NodeID:        0,
				CollectionIDs: []int{0},
			},
			testUtils.SubscribeToCollection{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"name":	"Fred",
					"age":	22
				}`,
			},
			testUtils.WaitForSync{},
			// both nodes should have the same results
			testUtils.Request{
				Request: `query {
					User {
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
						{
							"name": "Fred",
							"age":  int64(22),
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
						commits(fieldId: "C") {
							signature {
								type
							}
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"signature": map[string]any{
								"type": coreblock.SignatureTypeECDSA256K,
							},
						},
						{
							"signature": map[string]any{
								"type": coreblock.SignatureTypeEd25519,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
