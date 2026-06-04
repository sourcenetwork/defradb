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
			&action.UpdateDoc{
				Doc: `{
					"name":	"Fred"
				}`,
			},
			&action.Request{
				Request: `query {
						_commits {
							cid
							links {
								cid
							}
							heads {
								cid
							}
						}
					}`,
				NonOrderedResults: true,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": gomega.And(collectionUpdateCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": updateCid,
								},
							},
							"heads": []map[string]any{
								{
									"cid": collectionCreateCid,
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
							"heads": []map[string]any{},
						},
						{
							"cid":   gomega.And(ageCreateCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{
							"cid":   gomega.And(nameUpdateCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{
								{
									"cid": nameCreateCid,
								},
							},
						},
						{
							"cid":   gomega.And(nameCreateCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{
							"cid": gomega.And(updateCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": nameUpdateCid,
								},
							},
							"heads": []map[string]any{
								{
									"cid": createCid,
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
							"heads": []map[string]any{},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
