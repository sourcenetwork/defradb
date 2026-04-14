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

	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestEncryptedIndexNew_SchemaWithEncryptedIndex_ShouldNotHinderQuerying(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding an @encryptedIndex directive in the schema does not prevent normal document querying.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int @encryptedIndex
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `
					query  {
						User {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestEncryptedIndexNew_AfterAddRequest_ShouldNotHinderQuerying(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Adding an encrypted index after documents are already stored does not prevent normal querying.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.NewEncryptedIndex{
				FieldName: "age",
			},
			&action.Request{
				Request: `
					query  {
						User {
							name
							age
						}
					}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"name": "John",
							"age":  int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestEncryptedIndexNew_IfNonExistentFieldIsGiven_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Creating an encrypted index on a field that does not exist in the collection returns an error.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int
					}
				`,
			},
			testUtils.NewEncryptedIndex{
				FieldName:     "verified",
				ExpectedError: db.NewErrEncryptedIndexOnNonExistentField("verified").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestEncryptedIndexNew_IfIndexAlreadyExists_ShouldReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Attempting to create a duplicate encrypted index on the same field returns an already-exists error.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String 
						age: Int @encryptedIndex
					}
				`,
			},
			testUtils.NewEncryptedIndex{
				FieldName:     "age",
				ExpectedError: db.NewErrEncryptedIndexAlreadyExists("age").Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
