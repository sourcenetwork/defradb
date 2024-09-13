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

func TestSchemaUpdatesReplaceCollectionErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, replace collection",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// Replace Users with Book
				Patch: `
					[
						{
							"op": "replace", "path": "/Users", "value": {
								"Name": "Book",
								"Fields": [
									{"Name": "name", "Kind": 11}
								]
							}
						}
					]
				`,
				// WARNING: An error is still expected if/when we allow the adding of collections, as this also
				// implies that the "Users" collection is to be deleted.  Only once we support the adding *and*
				// removal of collections should this not error.
				ExpectedError: "adding schema via patch is not supported. Name: Book",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
