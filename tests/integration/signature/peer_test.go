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

package signature

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/crypto"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestDocSignature_WithPeersAndSecp256k1KeyType_ShouldSync(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		IdentityTypes: map[state.Identity]crypto.KeyType{
			testUtils.NodeIdentity(0).Value(): crypto.KeyTypeSecp256k1,
			testUtils.NodeIdentity(1).Value(): crypto.KeyTypeSecp256k1,
		},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
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
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
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
		IdentityTypes: map[state.Identity]crypto.KeyType{
			testUtils.NodeIdentity(0).Value(): crypto.KeyTypeEd25519,
			testUtils.NodeIdentity(1).Value(): crypto.KeyTypeEd25519,
		},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
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
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			testUtils.WaitForSync{},
			&action.Request{
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
		IdentityTypes: map[state.Identity]crypto.KeyType{
			testUtils.NodeIdentity(0).Value(): crypto.KeyTypeSecp256k1,
			testUtils.NodeIdentity(1).Value(): crypto.KeyTypeEd25519,
		},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
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
			testUtils.AddCollectionSubscription{
				NodeID:        0,
				CollectionIDs: []int{0},
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"name":	"Fred",
					"age":	22
				}`,
			},
			testUtils.WaitForSync{},
			// both nodes should have the same results
			&action.Request{
				Request: `query {
					User {
						name
						age
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "Fred",
							"age":  int64(22),
						},
						{
							"name": "John",
							"age":  int64(21),
						},
					},
				},
			},
			&action.Request{
				Request: `query {
						_commits(filter: {fieldName: {_eq: "_C"}}) {
							signature {
								type
								identity
							}
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeEd25519,
								"identity": newIdentityMatcher(testUtils.NodeIdentity(1).Value()),
							},
						},
						{
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": newIdentityMatcher(testUtils.NodeIdentity(0).Value()),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocSignature_WithPeersAnDifferentKeyTypesUpdatingSameDoc_ShouldSync(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		IdentityTypes: map[state.Identity]crypto.KeyType{
			testUtils.NodeIdentity(0).Value(): crypto.KeyTypeSecp256k1,
			testUtils.NodeIdentity(1).Value(): crypto.KeyTypeEd25519,
		},
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int
						verified: Boolean
					}
				`,
			},
			testUtils.ConnectPeers{
				SourceNodeID: 1,
				TargetNodeID: 0,
			},
			testUtils.AddCollectionSubscription{
				NodeID:        0,
				CollectionIDs: []int{0},
			},
			testUtils.AddCollectionSubscription{
				NodeID:        1,
				CollectionIDs: []int{0},
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: `{
					"verified": true
				}`,
			},
			testUtils.WaitForSync{},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"age": 23
				}`,
			},
			testUtils.WaitForSync{},
			// both nodes should have the same results
			&action.Request{
				Request: `query {
					User {
						name
						age
						verified
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name":     "John",
							"age":      int64(23),
							"verified": true,
						},
					},
				},
			},
			&action.Request{
				Request: `query {
						_commits(filter: {fieldName: {_eq: "_C"}}, order: {height: DESC}) {
							signature {
								type
								identity
							}
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": newIdentityMatcher(testUtils.NodeIdentity(0).Value()),
							},
						},
						{
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeEd25519,
								"identity": newIdentityMatcher(testUtils.NodeIdentity(1).Value()),
							},
						},
						{
							"signature": map[string]any{
								"type":     coreblock.SignatureTypeECDSA256K,
								"identity": newIdentityMatcher(testUtils.NodeIdentity(0).Value()),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
