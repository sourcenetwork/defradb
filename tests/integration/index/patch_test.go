// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestPatchCollection_AddSecondaryIndex_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						email: String
					}
				`,
			},
			testUtils.PatchCollection{
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
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						email: String @index
					}
				`,
			},
			testUtils.PatchCollection{
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
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						email: String @index
					}
				`,
			},
			testUtils.PatchCollection{
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
