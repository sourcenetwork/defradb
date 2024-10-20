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

func TestUniqueCompositeIndexUpdate_UponUpdatingDocWithExistingFieldValue_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Description: "updating non-indexed fields on a doc with existing field combination for composite index should succeed",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @index(unique: true, includes: [{field: "name"}, {field: "age"}]) {
						name: String 
						age: Int 
						email: String
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21,
						"email": "email@gmail.com"
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `
					{
						"email": "another@gmail.com"
					}`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
