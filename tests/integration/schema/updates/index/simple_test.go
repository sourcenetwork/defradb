// Copyright 2023 Democratized Data Foundation
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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestPatching_ForCollectionWithIndex_StillWorks(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test patching schema for collection with index still works",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String @index
						age:  Int    @index
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						age
						email
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestPatching_IfAttemptToAddIndex_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test adding index to collection via patch fails",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String @index
						age:  Int
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Indexes/-", "value": {
							"Name": "some_index",
							"ID": 0,
							"Fields": [
							  {
								"Name": "age",
								"Direction": "ASC"
							  }
							]
						  }
						}
					]
				`,
				ExpectedError: "adding indexes via patch is not supported. ProposedName: some_index",
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestPatching_IfAttemptToDropIndex_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test dropping index from collection via patch fails",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String @index
						age:  Int    @index(name: "users_age_index")
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "remove", "path": "/Users/Indexes/1" }
					]
				`,
				ExpectedError: "dropping indexes via patch is not supported. Name: users_age_index",
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestPatching_IfAttemptToChangeIndexName_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test changing index's name via patch fails",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String @index
						age:  Int    
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "replace", "path": "/Users/Indexes/0/Name", "value": "new_index_name" }
					]
				`,
				ExpectedError: "adding indexes via patch is not supported. ProposedName: new_index_name",
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestPatching_IfAttemptToChangeIndexField_ReturnError(t *testing.T) {
	testCases := []struct {
		description string
		patch       string
	}{
		{
			description: "Test adding a field to an index via patch fails",
			patch: `
					[
						{ "op": "add", "path": "/Users/Indexes/0/Fields/-", "value": {
							"Name": "age",
							"Direction": "ASC"
						  }
						}
					]
				`,
		},
		{
			description: "Test removing a field from an index via patch fails",
			patch: `
					[
						{ "op": "remove", "path": "/Users/Indexes/0/Fields/0" }
					]
				`,
		},
		{
			description: "Test changing index's field name via patch fails",
			patch: `
					[
						{ "op": "replace", "path": "/Users/Indexes/0/Fields/0/Name", "value": "new_field_name" }
					]
				`,
		},
		{
			description: "Test changing index's field direction via patch fails",
			patch: `
					[
						{ "op": "replace", "path": "/Users/Indexes/0/Fields/0/Direction", "value": "DESC" }
					]
				`,
		},
	}

	for _, testCase := range testCases {
		test := testUtils.TestCase{
			Description: testCase.description,
			Actions: []any{
				testUtils.SchemaUpdate{
					Schema: `
					type Users {
						name: String @index
						age:  Int    
					}
				`,
				},
				testUtils.SchemaPatch{
					Patch:         testCase.patch,
					ExpectedError: "changing indexes via patch is not supported",
				},
				testUtils.Request{
					Request: `query {
					Users {
						name
					}
				}`,
					Results: []map[string]any{},
				},
			},
		}
		testUtils.ExecuteTestCase(t, []string{"Users"}, test)
	}
}
