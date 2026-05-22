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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsOrderedAndGroupedByDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Fred",
						"age":	25
					}`,
			},
			&action.Request{
				Request: ` {
					_commits(groupBy: [docID], order: {docID: DESC}) {
						docID
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"docID": "{{.DocID0_1}}",
						},
						{
							"docID": "{{.DocID0_0}}",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
