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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestEncryptedIndexList_ShouldReturnListOfExistingIndexes(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @encryptedIndex
						age: Int @encryptedIndex
					}
					
					type Address {
						street: String @encryptedIndex
					}
				`,
			},
			testUtils.ListEncryptedIndexes{
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
			testUtils.ListEncryptedIndexes{
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

func TestEncryptedIndexList_IfIndexAddedLater_ShouldReturnListOfExistingIndexes(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @encryptedIndex
						age: Int 
					}
				`,
			},
			testUtils.ListEncryptedIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.EncryptedIndexDescription{
					{
						FieldName: "name",
						Type:      client.EncryptedIndexTypeEquality,
					},
				},
			},
			testUtils.NewEncryptedIndex{
				FieldName: "age",
			},
			testUtils.ListEncryptedIndexes{
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

func TestEncryptedIndexList_WhenRequestingAllIndexes_ShouldReturn(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @encryptedIndex
						age: Int @encryptedIndex
					}
					
					type Address {
						street: String @encryptedIndex
					}
				`,
			},
			testUtils.ListAllEncryptedIndexes{
				ExpectedIndexes: map[client.CollectionName][]client.EncryptedIndexDescription{
					"User": {
						{
							FieldName: "name",
							Type:      client.EncryptedIndexTypeEquality,
						},
						{
							FieldName: "age",
							Type:      client.EncryptedIndexTypeEquality,
						},
					},
					"Address": {
						{
							FieldName: "street",
							Type:      client.EncryptedIndexTypeEquality,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
