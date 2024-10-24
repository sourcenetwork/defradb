// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommits_AfterDocDeletion_ShouldStillFetch(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.DeleteDoc{
				DocID: 0,
			},
			testUtils.Request{
				Request: `
					query {
						commits(fieldId: "C") {
							cid
							fieldName
							links {
								cid
								name
							}
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":       testUtils.NewUniqueCid("delete"),
							"fieldName": nil,
							"links": []map[string]any{
								{
									"cid":  testUtils.NewUniqueCid("create composite"),
									"name": "_head",
								},
							},
						},
						{
							"cid":       testUtils.NewUniqueCid("create composite"),
							"fieldName": nil,
							"links": []map[string]any{
								{
									"cid":  testUtils.NewUniqueCid("create age"),
									"name": "age",
								},
								{
									"cid":  testUtils.NewUniqueCid("create name"),
									"name": "name",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
