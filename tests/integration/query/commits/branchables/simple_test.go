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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsBranchables(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
			&action.AddSchema{
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
							"schemaVersionId": "bafyreifnbhwntycylk2l6n4khiocdt3vks46tizjdaz6yx4tsmdjtdtlma",
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
							"schemaVersionId": "bafyreifnbhwntycylk2l6n4khiocdt3vks46tizjdaz6yx4tsmdjtdtlma",
							"delta":           testUtils.CBORValue(21),
							"docID":           "bae-f895da58-3326-510a-87f3-d043ff5424ea",
							"fieldName":       "age",
							"height":          int64(1),
							"links":           []map[string]any{},
						},
						{
							"cid":             gomega.And(nameCid, uniqueCid),
							"schemaVersionId": "bafyreifnbhwntycylk2l6n4khiocdt3vks46tizjdaz6yx4tsmdjtdtlma",
							"delta":           testUtils.CBORValue("John"),
							"docID":           "bae-f895da58-3326-510a-87f3-d043ff5424ea",
							"fieldName":       "name",
							"height":          int64(1),
							"links":           []map[string]any{},
						},
						{
							"cid":             gomega.And(compositeCid, uniqueCid),
							"schemaVersionId": "bafyreifnbhwntycylk2l6n4khiocdt3vks46tizjdaz6yx4tsmdjtdtlma",
							"delta":           nil,
							"docID":           "bae-f895da58-3326-510a-87f3-d043ff5424ea",
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
