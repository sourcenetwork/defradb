// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package encryption

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryption_WithEncryption_ShouldFetchDecrypted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
                    type Users {
                        name: String
                        age: Int
                    }
                `},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
				IsEncrypted: true,
			},
			testUtils.Request{
				Request: `
                    query {
                        Users {
                            _docID
                            name
                            age
                        }
                    }`,
				Results: []map[string]any{
					{
						"_docID": "bae-0b2f15e5-bfe7-5cb7-8045-471318d7dbc3",
						"name":   "John",
						"age":    int64(21),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
