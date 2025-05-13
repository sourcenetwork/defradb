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

	"github.com/onsi/gomega"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsBranchables(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

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
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
						{
							"cid": uniqueCid,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsBranchables_WithAllFields(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	collectionCid := testUtils.NewSameValue()
	compositeCid := testUtils.NewSameValue()
	ageCid := testUtils.NewSameValue()
	nameCid := testUtils.NewSameValue()

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
							schemaVersionId
							delta
							docID
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
							"cid":             gomega.And(collectionCid, uniqueCid),
							"schemaVersionId": "bafkreihhd6bqrjhl5zidwztgxzeseveplv3cj3fwtn3unjkdx7j2vr2vrq",
							"delta":           nil,
							"docID":           nil,
							"fieldName":       nil,
							"height":          int64(1),
							"links": []map[string]any{
								{
									"cid":  compositeCid,
									"name": nil,
								},
							},
						},
						{
							"cid":             gomega.And(ageCid, uniqueCid),
							"schemaVersionId": "bafkreihhd6bqrjhl5zidwztgxzeseveplv3cj3fwtn3unjkdx7j2vr2vrq",
							"delta":           testUtils.CBORValue(21),
							"docID":           "bae-0b2f15e5-bfe7-5cb7-8045-471318d7dbc3",
							"fieldName":       "age",
							"height":          int64(1),
							"links":           []map[string]any{},
						},
						{
							"cid":             gomega.And(nameCid, uniqueCid),
							"schemaVersionId": "bafkreihhd6bqrjhl5zidwztgxzeseveplv3cj3fwtn3unjkdx7j2vr2vrq",
							"delta":           testUtils.CBORValue("John"),
							"docID":           "bae-0b2f15e5-bfe7-5cb7-8045-471318d7dbc3",
							"fieldName":       "name",
							"height":          int64(1),
							"links":           []map[string]any{},
						},
						{
							"cid":             gomega.And(compositeCid, uniqueCid),
							"schemaVersionId": "bafkreihhd6bqrjhl5zidwztgxzeseveplv3cj3fwtn3unjkdx7j2vr2vrq",
							"delta":           nil,
							"docID":           "bae-0b2f15e5-bfe7-5cb7-8045-471318d7dbc3",
							"fieldName":       "_C",
							"height":          int64(1),
							"links": []map[string]any{
								{
									"cid":  ageCid,
									"name": "age",
								},
								{
									"cid":  nameCid,
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
