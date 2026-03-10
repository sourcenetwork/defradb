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
