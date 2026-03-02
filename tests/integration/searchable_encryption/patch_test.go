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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// TestPatchCollection_AddEncryptedIndex_ShouldError verifies that encrypted indexes cannot be added via patch.
// Since EncryptedIndexes is not exposed in the JSON representation, attempting to patch it
// results in an unmarshaling error, effectively preventing the mutation.
func TestPatchCollection_AddEncryptedIndex_ShouldError(t *testing.T) {
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
