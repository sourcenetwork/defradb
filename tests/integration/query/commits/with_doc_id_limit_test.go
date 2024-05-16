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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsWithDocIDAndLimit(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple all commits query with docID and limit",
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	23
				}`,
			},
			testUtils.Request{
				Request: ` {
						commits(docID: "bae-f54b9689-e06e-5e3a-89b3-f3aee8e64ca7", limit: 2) {
							cid
						}
					}`,
				Results: []map[string]any{
					{
						"cid": "bafyreial53rqep7uoheucc3rzvhs6dbhydnkbqe4w2bhd3hsybub4u3h6m",
					},
					{
						"cid": "bafyreictnkwvit6jp4mwhai3xp75nvtacxpq6zgbbjm55ylae3t6qshrze",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
