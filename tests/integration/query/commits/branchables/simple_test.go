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

func TestQueryCommitsBranchables(t *testing.T) {
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
			testUtils.Request{
				Request: `query {
						commits {
							cid
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid": testUtils.NewUniqueCid("collection"),
						},
						{
							"cid": testUtils.NewUniqueCid("name"),
						},
						{
							"cid": testUtils.NewUniqueCid("age"),
						},
						{
							"cid": testUtils.NewUniqueCid("head"),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsBranchables_WithAllFields(t *testing.T) {
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
			testUtils.Request{
				Request: `query {
						commits {
							cid
							collectionID
							delta
							docID
							fieldId
							fieldName
							height
							links {
								cid
								name
							}
						}
					}`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":          testUtils.NewUniqueCid("collection"),
							"collectionID": int64(1),
							"delta":        nil,
							"docID":        nil,
							"fieldId":      nil,
							"fieldName":    nil,
							"height":       int64(1),
							"links": []map[string]any{
								{
									"cid":  testUtils.NewUniqueCid("composite"),
									"name": nil,
								},
							},
						},
						{
							"cid":          testUtils.NewUniqueCid("age"),
							"collectionID": int64(1),
							"delta":        testUtils.CBORValue(21),
							"docID":        "bae-0b2f15e5-bfe7-5cb7-8045-471318d7dbc3",
							"fieldId":      "1",
							"fieldName":    "age",
							"height":       int64(1),
							"links":        []map[string]any{},
						},
						{
							"cid":          testUtils.NewUniqueCid("name"),
							"collectionID": int64(1),
							"delta":        testUtils.CBORValue("John"),
							"docID":        "bae-0b2f15e5-bfe7-5cb7-8045-471318d7dbc3",
							"fieldId":      "2",
							"fieldName":    "name",
							"height":       int64(1),
							"links":        []map[string]any{},
						},
						{
							"cid":          testUtils.NewUniqueCid("composite"),
							"collectionID": int64(1),
							"delta":        nil,
							"docID":        "bae-0b2f15e5-bfe7-5cb7-8045-471318d7dbc3",
							"fieldId":      "C",
							"fieldName":    nil,
							"height":       int64(1),
							"links": []map[string]any{
								{
									"cid":  testUtils.NewUniqueCid("age"),
									"name": "age",
								},
								{
									"cid":  testUtils.NewUniqueCid("name"),
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
