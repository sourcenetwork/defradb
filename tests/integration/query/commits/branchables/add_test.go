// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package branchables

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsBranchables_WithMultipleAdd(t *testing.T) {
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
			&action.AddCollection{
				SDL: `
					type Users @branchable {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"Fred",
					"age":	25
				}`,
			},
			&action.Request{
				Request: `query {
						_commits {
							cid
							heads {
								cid
							}
							links {
								cid
							}
						}
					}`,
				NonOrderedResults: true,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": gomega.And(doc2CollectionCid, uniqueCid),
							"heads": []map[string]any{
								{
									"cid": doc1CollectionCid,
								},
							},
							"links": []map[string]any{
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
							"heads": []map[string]any{},
						},
						{
							"cid":   gomega.And(doc2AgeFieldCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{
							"cid":   gomega.And(doc2NameFieldCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{},
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
							"heads": []map[string]any{},
						},
						{
							"cid":   gomega.And(doc1AgeFieldCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{
							"cid":   gomega.And(doc1NameFieldCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{},
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
							"heads": []map[string]any{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
