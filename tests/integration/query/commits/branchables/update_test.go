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

func TestQueryCommitsBranchables_WithDocUpdate(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	collectionUpdateCid := testUtils.NewSameValue()
	collectionCreateCid := testUtils.NewSameValue()
	updateCid := testUtils.NewSameValue()
	createCid := testUtils.NewSameValue()
	ageCreateCid := testUtils.NewSameValue()
	nameUpdateCid := testUtils.NewSameValue()
	nameCreateCid := testUtils.NewSameValue()

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
			testUtils.UpdateDoc{
				Doc: `{
					"name":	"Fred"
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
							"cid": gomega.And(collectionUpdateCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": collectionCreateCid,
								},
								{
									"cid": updateCid,
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
							"cid":   gomega.And(ageCreateCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid": gomega.And(nameUpdateCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": nameCreateCid,
								},
							},
						},
						{
							"cid":   gomega.And(nameCreateCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid": gomega.And(updateCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": createCid,
								},
								{
									"cid": nameUpdateCid,
								},
							},
						},
						{
							"cid": gomega.And(createCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": ageCreateCid,
								},
								{
									"cid": nameCreateCid,
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
