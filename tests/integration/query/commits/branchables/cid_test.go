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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestQueryCommitsBranchables_WithCidParam(t *testing.T) {
	test := testUtils.TestCase{
		// The collection-level commit CID is hardcoded in this test, and no
		// template placeholder exists for it yet.
		// See https://github.com/sourcenetwork/defradb/issues/4744.
		MultiplierExcludes: []string{multiplier.SignedDocs},
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
			&action.Request{
				Request: `query {
						_commits(
							cid: "bafyreignwrgxxwvuijnrjssobtd4qdzjdho2u2myumzthtcuukoo4txxjy"
						) {
							cid
							docID
							fieldName
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": "bafyreignwrgxxwvuijnrjssobtd4qdzjdho2u2myumzthtcuukoo4txxjy",
							// Extra params are used to verify this is a collection level cid
							"docID":     nil,
							"fieldName": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
