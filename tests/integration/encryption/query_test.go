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

func TestDocEncryption_WithEncryptionOnCounterCRDT_ShouldFetchDecrypted(t *testing.T) {
	const docID = "bae-ab8ae7d9-6473-5101-ba02-66b217948d7a"

	const query = `
		query {
			Users {
				_docID
				name
				points
			}
		}`

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
                    type Users {
                        name: String
                        points: Int @crdt(type: "pcounter")
                    }
                `},
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"points": 5
					}`,
				IsEncrypted: true,
			},
			testUtils.Request{
				Request: query,
				Results: []map[string]any{
					{
						"_docID": docID,
						"name":   "John",
						"points": 5,
					},
				},
			},
			testUtils.UpdateDoc{
				DocID: 0,
				Doc:   `{ "points": 3 }`,
			},
			testUtils.Request{
				Request: query,
				Results: []map[string]any{
					{
						"_docID": docID,
						"name":   "John",
						"points": 8,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
