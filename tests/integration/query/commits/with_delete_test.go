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

package commits

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommits_AfterDocDeletion_ShouldStillFetch(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	deleteCid := testUtils.NewSameValue()
	createCompositeCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
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
				Request: `
					query {
						_commits(filter: {fieldName: {_eq: "_C"}}) {
							cid
							fieldName
							links {
								cid
								fieldName
							}
							heads {
								cid
							}
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":       gomega.And(deleteCid, uniqueCid),
							"fieldName": "_C",
							"links":     []map[string]any{},
							"heads": []map[string]any{
								{
									"cid": createCompositeCid,
								},
							},
						},
						{
							"cid":       gomega.And(createCompositeCid, uniqueCid),
							"fieldName": "_C",
							"links": []map[string]any{
								{
									"cid":       testUtils.ValidCID(),
									"fieldName": "age",
								},
								{
									"cid":       testUtils.ValidCID(),
									"fieldName": "name",
								},
							},
							"heads": []map[string]any{},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
