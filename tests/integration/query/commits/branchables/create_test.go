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

func TestQueryCommitsBranchables_WithMultipleCreate(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	doc1NameFieldCid := testUtils.NewSameValue()
	doc1AgeFieldCid := testUtils.NewSameValue()
	doc2NameFieldCid := testUtils.NewSameValue()
	doc2AgeFieldCid := testUtils.NewSameValue()
	doc2CollectionCid := testUtils.NewSameValue()
	doc1CollectionCid := testUtils.NewSameValue()
	doc2CompositeCid := testUtils.NewSameValue()
	doc1CompositeCid := testUtils.NewSameValue()

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
							"cid": gomega.And(doc2CollectionCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": doc1CollectionCid,
								},
								{
									"cid": doc2CompositeCid,
								},
							},
						},
						{
							"cid": gomega.And(doc1CollectionCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": doc1CompositeCid,
								},
							},
						},
						{
							"cid":   gomega.And(doc2AgeFieldCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid":   gomega.And(doc2NameFieldCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid": gomega.And(doc2CompositeCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": doc2AgeFieldCid,
								},
								{
									"cid": doc2NameFieldCid,
								},
							},
						},
						{
							"cid":   gomega.And(doc1AgeFieldCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid":   gomega.And(doc1NameFieldCid, uniqueCid),
							"links": []map[string]any{},
						},
						{
							"cid": gomega.And(doc1CompositeCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": doc1AgeFieldCid,
								},
								{
									"cid": doc1NameFieldCid,
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
