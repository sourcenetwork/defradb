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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/internal/encryption"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryptionPeer_IfEncryptedDocHasIndexedField_ShouldIndexAfterDecryption(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int @index
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
						"name":	"Shahzad",
						"age":	25
					}`,
			},
			testUtils.CreateDoc{
				NodeID:         immutable.Some(0),
				Doc:            islam33Doc,
				IsDocEncrypted: true,
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
						"name":	"Andy",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				NodeID:         immutable.Some(0),
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			testUtils.WaitForSync{
				Event:   immutable.Some(encryption.KeysRetrievedEventName),
				NodeIDs: []int{1},
				Count:   2,
			},
			testUtils.Request{
				Request: `
					query @explain(type: execute) {
						User(filter: {age: {_eq: 21}}) {
							age
						}
					}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionPeer_IfDocDocHasEncryptedIndexedField_ShouldIndexAfterDecryption(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int @index
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
						"name":	"Shahzad",
						"age":	25
					}`,
			},
			testUtils.CreateDoc{
				NodeID:          immutable.Some(0),
				Doc:             islam33Doc,
				EncryptedFields: []string{"age"},
			},
			testUtils.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
						"name":	"Andy",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				NodeID:          immutable.Some(0),
				Doc:             john21Doc,
				EncryptedFields: []string{"age"},
			},
			testUtils.WaitForSync{
				Event:   immutable.Some(encryption.KeysRetrievedEventName),
				NodeIDs: []int{1},
				Count:   2,
			},
			testUtils.Request{
				Request: `
					query @explain(type: execute) {
						User(filter: {age: {_eq: 21}}) {
							age
						}
					}`,
				Asserter: testUtils.NewExplainAsserter().WithIndexFetches(2),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
