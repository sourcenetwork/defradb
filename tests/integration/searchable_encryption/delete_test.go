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
