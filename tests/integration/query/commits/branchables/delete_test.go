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
			testUtils.DeleteDoc{
				DocID: 0,
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
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": gomega.And(collectionDeleteCid, uniqueCid),
							"links": []map[string]any{
								{
									"cid": deleteCid,
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
							"cid":   gomega.And(nameCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{
							"cid":   gomega.And(ageCid, uniqueCid),
							"links": []map[string]any{},
							"heads": []map[string]any{},
						},
						{
							"cid":   gomega.And(deleteCid, uniqueCid),
							"links": []map[string]any{},
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
									"cid": nameCid,
								},
								{
									"cid": ageCid,
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
