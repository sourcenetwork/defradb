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
		Description: "Test schema update, add field",
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
		Description: "Test adding index to collection fails",
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
		Description: "Test adding index to collection fails",
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
