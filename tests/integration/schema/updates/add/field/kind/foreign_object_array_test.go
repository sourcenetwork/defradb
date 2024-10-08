// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kind

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesAddFieldKindForeignObjectArray_UnknownSchema(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array, unknown schema",
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo", "Kind": "[Unknown]"
						}}
					]
				`,
				ExpectedError: "no type found for given name. Field: foo, Kind: [Unknown]",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesAddFieldKindForeignObjectArray_KnownSchema(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array, known schema",
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {
							"Name": "foo", "Kind": "[Users]"
						}}
					]
				`,
				ExpectedError: "secondary relation fields cannot be defined on the schema. Name: foo",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
