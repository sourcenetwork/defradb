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

func TestQueryCommitsBranchables_WithMultipleCreate(t *testing.T) {
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
			testUtils.CreateDoc{
				Doc: `{
					"name":	"Fred",
					"age":	25
				}`,
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
							"cid": testUtils.NewUniqueCid("collection, doc2 create"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("collection, doc1 create"),
								},
								{
									"cid": testUtils.NewUniqueCid("doc2 create"),
								},
							},
						},
						{
							"cid": testUtils.NewUniqueCid("collection, doc1 create"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("doc1 create"),
								},
							},
						},
						{
							"cid":   testUtils.NewUniqueCid("doc1 name"),
							"links": []map[string]any{},
						},
						{
							"cid":   testUtils.NewUniqueCid("doc1 age"),
							"links": []map[string]any{},
						},
						{
							"cid": testUtils.NewUniqueCid("doc1 create"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("doc1 name"),
								},
								{
									"cid": testUtils.NewUniqueCid("doc1 age"),
								},
							},
						},
						{
							"cid":   testUtils.NewUniqueCid("doc2 name"),
							"links": []map[string]any{},
						},
						{
							"cid":   testUtils.NewUniqueCid("doc2 age"),
							"links": []map[string]any{},
						},
						{
							"cid": testUtils.NewUniqueCid("doc2 create"),
							"links": []map[string]any{
								{
									"cid": testUtils.NewUniqueCid("doc2 name"),
								},
								{
									"cid": testUtils.NewUniqueCid("doc2 age"),
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
