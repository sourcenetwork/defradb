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

package searchable_encryption

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryptionPeer_WithSimpleRequest_ShouldFetchSuccessfully(t *testing.T) {
	test := testUtils.TestCase{
		Description:                "A searchable encryption query on an encrypted-indexed field returns matching document IDs.",
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int @encryptedIndex
						verified: Boolean
					}`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSESync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						encrypted_User(filter: {age: {_eq: 21}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"encrypted_User": []map[string]any{
						{
							"docIDs": gomega.ConsistOf(testUtils.DocIDAt(0, 0)),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionPeer_WithMultipleEncryptedFields_QueryShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description:                "Searchable encryption queries on each of multiple encrypted-indexed fields all return the correct document ID.",
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User {
						name: String @encryptedIndex
						age: Int @encryptedIndex
						city: String @encryptedIndex
						verified: Boolean
					}`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John",
					"age": 25,
					"city": "New York",
					"verified": true
				}`,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSESync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						encrypted_User(filter: {name: {_eq: "John"}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"encrypted_User": []map[string]any{
						{
							"docIDs": gomega.ConsistOf(testUtils.DocIDAt(0, 0)),
						},
					},
				},
			},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						encrypted_User(filter: {age: {_eq: 25}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"encrypted_User": []map[string]any{
						{
							"docIDs": gomega.ConsistOf(testUtils.DocIDAt(0, 0)),
						},
					},
				},
			},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						encrypted_User(filter: {city: {_eq: "New York"}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"encrypted_User": []map[string]any{
						{
							"docIDs": gomega.ConsistOf(testUtils.DocIDAt(0, 0)),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionPeer_WithMultipleDocs_ShouldFilterCorrectly(t *testing.T) {
	test := testUtils.TestCase{
		Description:                "A searchable encryption query filters multiple encrypted documents by value and returns only matching IDs.",
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int @encryptedIndex
						verified: Boolean
					}`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
				IsDocEncrypted: true,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Alice",
					"age": 30
				}`,
				IsDocEncrypted: true,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Bob",
					"age": 30
				}`,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSESync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						encrypted_User(filter: {age: {_eq: 21}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"encrypted_User": []map[string]any{
						{
							"docIDs": gomega.ConsistOf(testUtils.DocIDAt(0, 0)),
						},
					},
				},
			},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						encrypted_User(filter: {age: {_eq: 30}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"encrypted_User": []map[string]any{
						{
							"docIDs": gomega.ConsistOf(
								testUtils.DocIDAt(0, 1),
								testUtils.DocIDAt(0, 2),
							),
						},
					},
				},
			},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						encrypted_User(filter: {age: {_eq: 33}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"encrypted_User": []map[string]any{
						{
							"docIDs": []string{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_IfThereIsNoIndex_EncryptedQueryShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Description:                "Issuing a searchable encryption query on a collection without any encrypted index returns a GraphQL error.",
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age: Int 
					}`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
				IsDocEncrypted: true,
			},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						encrypted_User(filter: {age: {_eq: 21}}) {
							docIDs
						}
					}`,
				ExpectedError: "Cannot query field \"encrypted_User\" on type \"Query\".",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_IfThereIsIndexButOnAnotherField_EncryptedQueryShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Description:                "Filtering a searchable encryption query on a non-indexed field returns an invalid argument error.",
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User {
						name: String @encryptedIndex
						age: Int 
					}`,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
				IsDocEncrypted: true,
			},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						encrypted_User(filter: {age: {_eq: 21}}) {
							docIDs
						}
					}`,
				ExpectedError: "Argument \"filter\" has invalid value",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionPeer_WithQueryOnMultipleFields_ShouldReturnIntersection(t *testing.T) {
	test := testUtils.TestCase{
		Description:                "A searchable encryption query filtering on multiple encrypted fields returns only the intersecting document IDs.",
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddCollection{
				SDL: `
					type User {
						name: String @encryptedIndex
						age: Int @encryptedIndex
						city: String @encryptedIndex
						verified: Boolean
					}`,
			},
			testUtils.AddReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
				IsDocEncrypted: true,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John",
					"age": 30
				}`,
				IsDocEncrypted: true,
			},
			&action.AddDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Bob",
					"age": 30
				}`,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSESync{},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						encrypted_User(filter: {name: {_eq: "John"}, age: {_eq: 30}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"encrypted_User": []map[string]any{
						{
							"docIDs": gomega.ConsistOf(testUtils.DocIDAt(0, 1)),
						},
					},
				},
			},
			&action.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						encrypted_User(filter: {name: {_eq: "Bob"}, age: {_eq: 21}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"encrypted_User": []map[string]any{
						{
							"docIDs": gomega.BeEmpty(),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
