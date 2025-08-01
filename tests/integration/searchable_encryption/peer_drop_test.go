// Copyright 2025 Democratized Data Foundation
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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryptionPeer_AfterDroppingIndex_ShouldReturnEmptyResults(t *testing.T) {
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
			testUtils.DropEncryptedIndex{
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				FieldName:    "age",
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
				ExpectedError: "invalid state, required property is uninitialized. Host: SelectEncrypted, PropertyName: collection has no encrypted indexes",
			},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `
					query {
						Users_encrypted(filter: {age: {_eq: 30}}) {
							docIDs
						}
					}`,
				ExpectedError: "invalid state, required property is uninitialized. Host: SelectEncrypted, PropertyName: collection has no encrypted indexes",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionPeer_AfterDroppingIndex_NewDocumentsShouldNotBeSearchable(t *testing.T) {
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
			testUtils.DropEncryptedIndex{
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				FieldName:    "age",
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name": "Bob",
					"age": 25
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
						Users_encrypted(filter: {age: {_eq: 25}}) {
							docIDs
						}
					}`,
				ExpectedError: "invalid state, required property is uninitialized. Host: SelectEncrypted, PropertyName: collection has no encrypted indexes",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
