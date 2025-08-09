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
	"github.com/sourcenetwork/defradb/internal/db"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestEncryptedIndexDelete_WithExistingIndex_ShouldDeleteSuccessfully(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int @encryptedIndex
					}
				`,
			},
			testUtils.GetEncryptedIndexes{
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
			testUtils.GetEncryptedIndexes{
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
			testUtils.SchemaUpdate{
				Schema: `
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

func TestEncryptedIndexDelete_AfterDelete_CanCreateIndexAnew(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int @encryptedIndex
					}
				`,
			},
			testUtils.DeleteEncryptedIndex{
				FieldName: "age",
			},
			testUtils.GetEncryptedIndexes{
				CollectionID:    0,
				ExpectedIndexes: []client.EncryptedIndexDescription{},
			},
			testUtils.CreateEncryptedIndex{
				FieldName: "age",
			},
			testUtils.GetEncryptedIndexes{
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
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @encryptedIndex
						age: Int @encryptedIndex
						city: String @encryptedIndex
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
					{
						FieldName: "city",
						Type:      client.EncryptedIndexTypeEquality,
					},
				},
			},
			testUtils.DeleteEncryptedIndex{
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
						FieldName: "city",
						Type:      client.EncryptedIndexTypeEquality,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
