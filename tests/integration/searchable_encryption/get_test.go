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

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestEncryptedIndexGet_ShouldReturnListOfExistingIndexes(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @encryptedIndex
						age: Int @encryptedIndex
					}
					
					type Address {
						street: String @encryptedIndex
					}
				`,
			},
			testUtils.GetEncryptedIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.EncryptedIndexDescription{
					{
						FieldName: "name",
						Type:      client.EncryptedIndexTypeEquality,
					},
					{
						FieldName: "age",
						Type:      client.EncryptedIndexTypeEquality,
					},
				},
			},
			testUtils.GetEncryptedIndexes{
				CollectionID: 1,
				ExpectedIndexes: []client.EncryptedIndexDescription{
					{
						FieldName: "street",
						Type:      client.EncryptedIndexTypeEquality,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestEncryptedIndexGet_IfIndexCreatedLater_ShouldReturnListOfExistingIndexes(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @encryptedIndex
						age: Int 
					}
				`,
			},
			testUtils.GetEncryptedIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.EncryptedIndexDescription{
					{
						FieldName: "name",
						Type:      client.EncryptedIndexTypeEquality,
					},
				},
			},
			testUtils.CreateEncryptedIndex{
				FieldName: "age",
			},
			testUtils.GetEncryptedIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.EncryptedIndexDescription{
					{
						FieldName: "name",
						Type:      client.EncryptedIndexTypeEquality,
					},
					{
						FieldName: "age",
						Type:      client.EncryptedIndexTypeEquality,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
