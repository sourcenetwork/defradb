// Copyright 2024 Democratized Data Foundation
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

func TestUniqueIndexUpdate_UponUpdatingDocNonIndexedField_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "updating non-indexed fields on a doc with a unique index should succeed",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String @index(unique: true)
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Fred",
						"age":	36
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `
					{
						"age":	37
					}`,
				ExpectedError: "can not index a doc's field(s) that violates unique index",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
