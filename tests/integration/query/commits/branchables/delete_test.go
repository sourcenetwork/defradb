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

func TestQueryCommitsBranchables_WithDelete(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	collectionDeleteCid := testUtils.NewSameValue()
	collectionCreateCid := testUtils.NewSameValue()
	deleteCid := testUtils.NewSameValue()
	createCid := testUtils.NewSameValue()
	nameCid := testUtils.NewSameValue()
	ageCid := testUtils.NewSameValue()

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
							"cid": gomega.And(collectionDeleteCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": collectionCreateCid,
								},
								{
									"cid": deleteCid,
								},
							},
						},
						{
							"cid": gomega.And(collectionCreateCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": createCid,
								},
							},
						},
						{
							"cid":   gomega.And(nameCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid":   gomega.And(ageCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid": gomega.And(deleteCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": createCid,
								},
							},
						},
						{
							"cid": gomega.And(createCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": nameCid,
								},
								{
									"cid": ageCid,
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
