// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package replace

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesReplaceFieldErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, replace field",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "replace", "path": "/Users/Schema/Fields/2", "value": {"Name": "Fax", "Kind": 11} }
					]
				`,
				ExpectedError: "deleting an existing field is not supported. Name: name, ID: 2",
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}

func TestSchemaUpdatesReplaceFieldWithIDErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, replace field with correct ID",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "replace", "path": "/Users/Schema/Fields/2", "value": {"ID":2, "Name": "fax", "Kind": 11} }
					]
				`,
				ExpectedError: "mutating an existing field is not supported. ID: 2, ProposedName: fax",
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
