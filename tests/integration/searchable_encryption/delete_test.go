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
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestEncryptedIndexDelete_WithExistingIndex_ShouldDeleteSuccessfully(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Deleting an existing encrypted index removes it from the collection's index list.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int @encryptedIndex
					}
				`,
			},
			testUtils.ListEncryptedIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.EncryptedIndexDescription{
					{
						FieldName: "age",
						Type:      client.EncryptedIndexTypeEquality,
					},
				},
			},
			testUtils.DeleteEncryptedIndex{
				FieldName: "age",
			},
			testUtils.ListEncryptedIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.EncryptedIndexDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestEncryptedIndexDelete_IfIndexDoesNotExist_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Deleting an encrypted index that does not exist returns a does-not-exist error.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			testUtils.DeleteEncryptedIndex{
				FieldName:     "age",
				ExpectedError: db.NewErrEncryptedIndexDoesNotExist("age").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestEncryptedIndexDelete_AfterDelete_CanMakeNewIndexAnew(t *testing.T) {
	test := testUtils.TestCase{
		Description: "After deleting an encrypted index, a new encrypted index can be created on the same field.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int @encryptedIndex
					}
				`,
			},
			testUtils.DeleteEncryptedIndex{
				FieldName: "age",
			},
			testUtils.ListEncryptedIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.EncryptedIndexDescription{},
			},
			testUtils.NewEncryptedIndex{
				FieldName: "age",
			},
			testUtils.ListEncryptedIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.EncryptedIndexDescription{
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

func TestEncryptedIndexDelete_MultipleIndexes_ShouldOnlyDeleteSpecified(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Deleting one encrypted index from a collection with multiple indexes leaves the others intact.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String @encryptedIndex
						age: Int @encryptedIndex
						city: String @encryptedIndex
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
					{
						FieldName: "city",
						Type:      client.EncryptedIndexTypeEquality,
					},
				},
			},
			testUtils.DeleteEncryptedIndex{
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
						FieldName: "city",
						Type:      client.EncryptedIndexTypeEquality,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
