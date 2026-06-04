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

// TestPatchCollection_NewEncryptedIndex_ShouldError verifies that encrypted indexes cannot be added via patch.
func TestPatchCollection_NewEncryptedIndex_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "add",
							"path": "/Users/EncryptedIndexes",
							"value": [
								{
									"FieldName": "email",
									"Type": "equality"
								}
							]
						}
					]
				`,
				ExpectedError: db.ErrCollectionEncryptedIndexesCannotBeMutated.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestPatchCollection_RemoveEncryptedIndex_ShouldError verifies that encrypted indexes cannot be removed via patch.
func TestPatchCollection_RemoveEncryptedIndex_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						email: String @encryptedIndex
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/Users/EncryptedIndexes"
						}
					]
				`,
				ExpectedError: db.ErrCollectionEncryptedIndexesCannotBeMutated.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestPatchCollection_ModifyEncryptedIndex_ShouldError verifies that encrypted indexes cannot be modified via patch.
func TestPatchCollection_ModifyEncryptedIndex_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						email: String @encryptedIndex
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/Users/EncryptedIndexes",
							"value": [
								{
									"FieldName": "name",
									"Type": "equality"
								}
							]
						}
					]
				`,
				ExpectedError: db.ErrCollectionEncryptedIndexesCannotBeMutated.Error(),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
