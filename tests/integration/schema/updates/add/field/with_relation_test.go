// Copyright 2024 Democratized Data Foundation
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

// This test ensures that nearby relation fields are not failing validation during a schema patch.
func TestSchemaUpdatesAddField_DoesNotAffectExistingRelation(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						author: Author
					}

					type Author {
						name: String
						books: [Book]
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Book/Fields/-", "value": {"Name": "rating", "Kind": 4} }
					]
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
