// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesTestAddField(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, passing test allows new field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "test", "path": "/Users/Schema/Name", "value": "Users" },
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						Email
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaUpdatesTestAddFieldBlockedByTest(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, failing test blocks new field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "test", "path": "/Users/Schema/Name", "value": "Author" },
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "Email", "Kind": 11} }
					]
				`,
				ExpectedError: "test failed",
			},
			testUtils.Request{
				Request: `query {
					Users {
						Name
						Email
					}
				}`,
				ExpectedError: "Cannot query field \"Email\" on type \"Users\"",
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
