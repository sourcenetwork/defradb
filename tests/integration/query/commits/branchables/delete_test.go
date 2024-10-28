// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package branchables

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsBranchables_WithDelete(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users @branchable {
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
				Request: `query {
						commits {
							cid
							links {
								cid
							}
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": testUtils.NewUniqueCid("collection, delete"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("collection, create"),
								},
								{
									"cid": testUtils.NewUniqueCid("delete"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("collection, create"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("create"),
								},
							},
						},
						{
							"cid":   testUtils.NewUniqueCid("name"),
							"links": []map[string]any{},
						},
						{
							"cid":   testUtils.NewUniqueCid("age"),
							"links": []map[string]any{},
						},
						{
							"cid": testUtils.NewUniqueCid("delete"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("create"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("create"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("name"),
								},
								{
									"cid": testUtils.NewUniqueCid("age"),
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
