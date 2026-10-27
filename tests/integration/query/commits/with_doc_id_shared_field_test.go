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

// Two documents that share an identical field value produce a byte-identical genesis field
// block (field deltas carry no DocID), so that block's CID is owned by both documents. A
// commits query must still return each document's complete, correctly-attributed history and
// must not drop the shared field commit from either owner.
func TestQueryCommitsWithDocID_SharedFieldBlockAcrossDocuments(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			// Same name (shared "name" field block), different age (unique "age" field block).
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Shared",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Shared",
						"age":	22
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(docID: "{{.DocID0_0}}") {
							fieldName
							docID
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"fieldName": "name", "docID": "{{.DocID0_0}}"},
						{"fieldName": "age", "docID": "{{.DocID0_0}}"},
						{"fieldName": "_C", "docID": "{{.DocID0_0}}"},
					},
				},
				NonOrderedResults: true,
			},
			&action.Request{
				Request: `query {
						_commits(docID: "{{.DocID0_1}}") {
							fieldName
							docID
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"fieldName": "name", "docID": "{{.DocID0_1}}"},
						{"fieldName": "age", "docID": "{{.DocID0_1}}"},
						{"fieldName": "_C", "docID": "{{.DocID0_1}}"},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
