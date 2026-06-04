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

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestPatchCollection_AddSecondaryIndex_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
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
							"path": "User/Indexes/-",
							"value": {
                                "Fields": [
                                    {
                                        "Name": "email",
                                        "Descending": false
                                    }
                                ],
                                "Unique": false
                            }
						}
					]
				`,
				ExpectedError: "collection indexes cannot be mutated",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPatchCollection_RemoveSecondaryIndex_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						email: String @index
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "User/Indexes"
						}
					]
				`,
				ExpectedError: "collection indexes cannot be mutated",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestPatchCollection_ModifySecondaryIndex_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						email: String @index
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "User/Indexes",
							"value": [
								{
									"Name": "User_name_ASC",
									"ID": 2,
									"Fields": [
										{
											"Name": "name",
											"Descending": false
										}
									],
									"Unique": false
								}
							]
						}
					]
				`,
				ExpectedError: "collection indexes cannot be mutated",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
