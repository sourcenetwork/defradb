// Copyright 2023 Democratized Data Foundation
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

func TestSchemaUpdatesAddFieldKindForeignObjectArray(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, add field with kind foreign object array (17)",
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
						{ "op": "add", "path": "/Users/Schema/Fields/-", "value": {"Name": "foo", "Kind": 17} }
					]
				`,
				ExpectedError: "the adding of new relation fields is not yet supported. Field: foo, Kind: 17",
			},
		},
	}
	testUtils.ExecuteTestCase(t, []string{"Users"}, test)
}
