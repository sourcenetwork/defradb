// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package field

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesMoveFieldErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, move field",
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
						{ "op": "move", "from": "/Users/Schema/Fields/1", "path": "/Users/Schema/Fields/-" }
					]
				`,
				ExpectedError: "moving fields is not currently supported. Name: name, ProposedIndex: 1, ExistingIndex: 2",
			},
		},
	}
	testUtils.ExecuteTEMP(t, test)
}
