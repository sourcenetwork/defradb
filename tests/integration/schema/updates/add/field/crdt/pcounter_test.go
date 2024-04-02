// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crdt

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdates_AddFieldCRDTPCounter_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with crdt P Counter (5)",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 4, "Typ": 5} }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						foo
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdates_AddFieldCRDTPCounterWithMismatchKind_Error(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with crdt P Counter (5)",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 2, "Typ": 5} }
					]
				`,
				ExpectedError: "CRDT type pcounter can't be assigned to field kind Boolean",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
