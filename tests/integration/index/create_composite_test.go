// Copyright 2023 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCompositeIndexCreate_WhenCreated_CanRetrieve(t *testing.T) {
	test := testUtils.TestCase{
		Description: "create composite index and retrieve it",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String 
						age: Int 
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `
					{
						"name":	"Shahzad",
						"age":	22
					}`,
			},
			testUtils.CreateIndex{
				CollectionID: 0,
				IndexName:    "name_age_index",
				FieldsNames:  []string{"name", "age"},
			},
			testUtils.GetIndexes{
				CollectionID: 0,
				ExpectedIndexes: []client.IndexDescription{
					{
						Name: "name_age_index",
						ID:   1,
						Fields: []client.IndexedFieldDescription{
							{
								Name:      "name",
								Direction: client.Ascending,
							},
							{
								Name:      "age",
								Direction: client.Ascending,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
