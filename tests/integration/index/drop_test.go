// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndexDrop_IfIndexDoesNotExist_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Drop index should return error if index does not exist",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Age: Int
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				// bae-52b9170d-b77a-5887-b877-cbdbb99b009f
				Doc: `
					{
						"Name":	"John",
						"Age":	21
					}`,
			},
			testUtils.DropIndex{
				CollectionID:  0,
				IndexName:     "non_existing_index",
				ExpectedError: "index with name doesn't exists. Name: non_existing_index",
			},
			testUtils.Request{
				Request: `
					query  {
						Users {
							Name
							Age
						}
					}`,
				Results: []map[string]any{
					{
						"Name": "John",
						"Age":  uint64(21),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
