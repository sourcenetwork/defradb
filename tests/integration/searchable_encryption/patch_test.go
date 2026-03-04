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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestPatchCollection_NewEncryptedIndex_ShouldError verifies that encrypted indexes cannot be added via patch.
// Since EncryptedIndexes is not exposed in the JSON representation, attempting to patch it
// results in an unmarshaling error, effectively preventing the mutation.
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
							"path": "/EncryptedIndexes",
							"value": [
								{
									"FieldName": "email",
									"Type": 0
								}
							]
						}
					]
				`,
				ExpectedError: "cannot unmarshal array into Go value",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestPatchCollection_RemoveEncryptedIndex_ShouldError verifies that encrypted indexes cannot be removed via patch.
// Since EncryptedIndexes is not exposed in the JSON representation, attempting to remove it
// results in an error about removing a nonexistent key, effectively preventing the mutation.
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
							"path": "/EncryptedIndexes"
						}
					]
				`,
				ExpectedError: "unable to remove nonexistent key",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestPatchCollection_ModifyEncryptedIndex_ShouldError verifies that encrypted indexes cannot be modified via patch.
// Since EncryptedIndexes is not exposed in the JSON representation, attempting to replace it
// results in an error about a missing key, effectively preventing the mutation.
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
							"path": "/EncryptedIndexes",
							"value": [
								{
									"FieldName": "name",
									"Type": 0
								}
							]
						}
					]
				`,
				ExpectedError: "doc is missing key",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
