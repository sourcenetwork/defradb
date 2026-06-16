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
		// See branchableCollectionCidExcludes (with_cid_branchable_test.go) — pending
		// https://github.com/sourcenetwork/defradb/issues/4744.
		MultiplierExcludes: branchableCollectionCidExcludes,
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
			&action.UpdateDoc{
				Doc: `{
					"name": "Freddddd"
				}`,
			},
			&action.Request{
				// This is the cid of the collection-commit when the second doc (John) is created.
				// Without the docID param both John and Fred should be returned.
				Request: `query {
					Users (
							cid: "{{.CollectionCID0_1}}",
							docID: "{{.DocID0_0}}"
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
