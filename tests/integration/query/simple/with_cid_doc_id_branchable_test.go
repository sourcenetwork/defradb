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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithCidOfBranchableCollectionAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users @branchable {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Fred"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Freddddd"
				}`,
			},
			&action.Request{
				// This is the cid of the collection-commit when the second doc (John) is created.
				// Without the docID param both John and Fred should be returned.
				Request: `query {
					Users (
							cid: "bafyreicxlkbypv4hjc2trunobsugx63no47322xfjjazj3v5y6hqeck3d4",
							docID: "bae-235c64e3-abf7-549c-9aff-971c8afdfa3f"
						) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
