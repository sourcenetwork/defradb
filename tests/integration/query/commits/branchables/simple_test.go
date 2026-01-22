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
			&action.Request{
				Request: `query {
						_commits {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
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
			&action.Request{
				Request: `query {
						_commits {
							cid
							collectionVersionId
							delta
							docID
							fieldName
							height
							links {
								cid
								fieldName
							}
							heads {
								cid
							}
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":                 gomega.And(collectionCid, uniqueCid),
							"collectionVersionId": "bafyreihsneodeja4lfer5puptim3lkwvketyckrmkhfpgxm67ch5wenjwq",
							"delta":               nil,
							"docID":               nil,
							"fieldName":           nil,
							"height":              int64(1),
							"links": []map[string]any{
								{
									"cid":       compositeCid,
									"fieldName": "_C",
								},
							},
							"heads": []map[string]any{},
						},
						{
							"cid":                 gomega.And(ageCid, uniqueCid),
							"collectionVersionId": "bafyreihsneodeja4lfer5puptim3lkwvketyckrmkhfpgxm67ch5wenjwq",
							"delta":               testUtils.CBORValue(21),
							"docID":               "bae-c65ccba7-7d6c-55c8-9d46-e865305f7790",
							"fieldName":           "age",
							"height":              int64(1),
							"links":               []map[string]any{},
							"heads":               []map[string]any{},
						},
						{
							"cid":                 gomega.And(nameCid, uniqueCid),
							"collectionVersionId": "bafyreihsneodeja4lfer5puptim3lkwvketyckrmkhfpgxm67ch5wenjwq",
							"delta":               testUtils.CBORValue("John"),
							"docID":               "bae-c65ccba7-7d6c-55c8-9d46-e865305f7790",
							"fieldName":           "name",
							"height":              int64(1),
							"links":               []map[string]any{},
							"heads":               []map[string]any{},
						},
						{
							"cid":                 gomega.And(compositeCid, uniqueCid),
							"collectionVersionId": "bafyreihsneodeja4lfer5puptim3lkwvketyckrmkhfpgxm67ch5wenjwq",
							"delta":               nil,
							"docID":               "bae-c65ccba7-7d6c-55c8-9d46-e865305f7790",
							"fieldName":           "_C",
							"height":              int64(1),
							"links": []map[string]any{
								{
									"cid":       ageCid,
									"fieldName": "age",
								},
								{
									"cid":       nameCid,
									"fieldName": "name",
								},
							},
							"heads": []map[string]any{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
