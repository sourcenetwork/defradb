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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryptionPeer_AfterDeletingIndex_SEQueryShouldReturnError(t *testing.T) {
	test := testUtils.TestCase{
		KMS:                        testUtils.KMS{Activated: true},
		EnableSearchableEncryption: true,
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						age: Int @encryptedIndex
						verified: Boolean
					}`,
			},
			testUtils.CreateReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			&action.CreateDoc{
				NodeID: immutable.Some(0),
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
				IsDocEncrypted: true,
			},
			testUtils.DeleteEncryptedIndex{
				NodeID:       immutable.Some(0),
				CollectionID: 0,
				FieldName:    "age",
			},
			testUtils.Wait{
				Duration: time.Millisecond * 100,
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
