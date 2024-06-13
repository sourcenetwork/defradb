// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package copy

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesCopyCollectionWithRemoveIDAndReplaceName(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, copy collection, rename and remove ids",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// Here we esentially use Users as a template, copying it and renaming the
				// clone. It is deliberately blocked for now, but should function at somepoint.
				Patch: `
					[
						{ "op": "copy", "from": "/Users", "path": "/Book" },
						{ "op": "remove", "path": "/Book/Root" },
						{ "op": "remove", "path": "/Book/VersionID" },
						{ "op": "replace", "path": "/Book/Name", "value": "Book" }
					]
				`,
				ExpectedError: "adding schema via patch is not supported. Name: Book",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
