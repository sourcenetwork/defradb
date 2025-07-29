// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package searchable_encryption

import (
	"testing"
	"time"

	"github.com/onsi/gomega"
	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryptionPeer_UponSync_ShouldSyncEncryptedDAG(t *testing.T) {
	test := testUtils.TestCase{
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						age: Int @encryptedIndex
						verified: Boolean
					}`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
				SEEnabled:    true,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
				IsDocEncrypted: true,
			},
			testUtils.Wait{
				Duration: time.Millisecond * 100,
			},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						Users_encrypted(filter: {age: {_eq: 21}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"Users_encrypted": []map[string]any{
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

func TestDocEncryptionPeer_WithMultipleEncryptedFields_ShouldSyncAllFields(t *testing.T) {
	test := testUtils.TestCase{
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String @encryptedIndex
						age: Int @encryptedIndex
						city: String @encryptedIndex
						verified: Boolean
					}`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
				SEEnabled:    true,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "John",
					"age": 25,
					"city": "New York",
					"verified": true
				}`,
				IsDocEncrypted: true,
			},
			testUtils.Wait{
				Duration: time.Millisecond * 100,
			},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						Users_encrypted(filter: {name: {_eq: "John"}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"Users_encrypted": []map[string]any{
						{
							"docIDs": gomega.ConsistOf(testUtils.DocIDAt(0, 0)),
						},
					},
				},
			},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						Users_encrypted(filter: {age: {_eq: 25}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"Users_encrypted": []map[string]any{
						{
							"docIDs": gomega.ConsistOf(testUtils.DocIDAt(0, 0)),
						},
					},
				},
			},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						Users_encrypted(filter: {city: {_eq: "New York"}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"Users_encrypted": []map[string]any{
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
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						age: Int @encryptedIndex
						verified: Boolean
					}`,
			},
			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
				SEEnabled:    true,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
				IsDocEncrypted: true,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Alice",
					"age": 30
				}`,
				IsDocEncrypted: true,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Bob",
					"age": 30
				}`,
				IsDocEncrypted: true,
			},
			testUtils.Wait{
				Duration: time.Millisecond * 100,
			},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						Users_encrypted(filter: {age: {_eq: 21}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"Users_encrypted": []map[string]any{
						{
							"docIDs": gomega.ConsistOf(testUtils.DocIDAt(0, 0)),
						},
					},
				},
			},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						Users_encrypted(filter: {age: {_eq: 30}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"Users_encrypted": []map[string]any{
						{
							"docIDs": gomega.ConsistOf(
								testUtils.DocIDAt(0, 1),
								testUtils.DocIDAt(0, 2),
							),
						},
					},
				},
			},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						Users_encrypted(filter: {age: {_eq: 33}}) {
							docIDs
						}
					}`,
				Results: map[string]any{
					"Users_encrypted": []map[string]any{
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
