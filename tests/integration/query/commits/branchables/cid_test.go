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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsBranchables_WithCidParam(t *testing.T) {
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
