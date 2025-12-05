// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
	createAgeCid := testUtils.NewSameValue()
	createNameCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
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
									"cid":       createAgeCid,
									"fieldName": "age",
								},
								{
									"cid":       createNameCid,
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
