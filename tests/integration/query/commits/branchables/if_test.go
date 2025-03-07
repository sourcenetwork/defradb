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

func TestQueryCommitsBranchables_WithIfDirectiveTrue(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	collectionCid := testUtils.NewSameValue()
	nameCid := testUtils.NewSameValue()
	ageCid := testUtils.NewSameValue()
	headCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users @branchable(if: true) {
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
							"cid": gomega.And(collectionCid, uniqueCid),
						},
						{
							"cid": gomega.And(nameCid, uniqueCid),
						},
						{
							"cid": gomega.And(ageCid, uniqueCid),
						},
						{
							"cid": gomega.And(headCid, uniqueCid),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsBranchables_WithIfDirectiveFalse(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	nameCid := testUtils.NewSameValue()
	ageCid := testUtils.NewSameValue()
	headCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users @branchable(if: false) {
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
						// Note: This collection is not branchable, there is no collection
						// level commit
						{
							"cid": gomega.And(nameCid, uniqueCid),
						},
						{
							"cid": gomega.And(ageCid, uniqueCid),
						},
						{
							"cid": gomega.And(headCid, uniqueCid),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
