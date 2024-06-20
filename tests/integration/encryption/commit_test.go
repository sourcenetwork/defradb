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

	"github.com/sourcenetwork/defradb/internal/encryption"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func encrypt(plaintext []byte) []byte {
	val, _ := encryption.EncryptAES(plaintext, []byte("examplekey1234567890examplekey12"))
	return val
}

func TestDocEncryption_ShouldStoreCommitsDeltaEncrypted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
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
						commits {
							cid
							collectionID
							delta
							docID
							fieldId
							fieldName
							height
							links {
								cid
								name
							}
						}
					}
				`,
				Results: []map[string]any{
					{
						"cid":          "bafyreidrbl46bz5nuzuby6s4zqvzliq4gyup3pq6ipy7ljm5o7l5hxtjhm",
						"collectionID": int64(1),
						"delta":        encrypt(testUtils.CBORValue(21)),
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreighzsctnwzhw57nbzici6dbvohozwet5w2baey3p4dxtxp7wxybui",
						"collectionID": int64(1),
						"delta":        encrypt(testUtils.CBORValue("John")),
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "2",
						"fieldName":    "name",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreidzfgvlx6eaj4furwl3mpvxp3wslbvzs4hvknivhpjw7g275k5v5i",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(1),
						"links": []map[string]any{
							{
								"cid":  "bafyreidrbl46bz5nuzuby6s4zqvzliq4gyup3pq6ipy7ljm5o7l5hxtjhm",
								"name": "age",
							},
							{
								"cid":  "bafyreighzsctnwzhw57nbzici6dbvohozwet5w2baey3p4dxtxp7wxybui",
								"name": "name",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_UponUpdate_ShouldEncryptedCommitDelta(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
				IsEncrypted: true,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits(fieldId: "1") {
							delta
						}
					}
				`,
				Results: []map[string]any{
					{
						"delta": encrypt(testUtils.CBORValue(22)),
					},
					{
						"delta": encrypt(testUtils.CBORValue(21)),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_WithMultipleDocsUponUpdate_ShouldEncryptedOnlyRelevantDocs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				// bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
				IsEncrypted: true,
			},
			testUtils.CreateDoc{
				// bae-d55bd956-1cc4-5d26-aa71-b98807ad49d6
				Doc: `{
						"name":	"Islam",
						"age":	33
					}`,
			},
			testUtils.UpdateDoc{
				DocID: 0,
				Doc: `{
					"age": 22
				}`,
			},
			testUtils.UpdateDoc{
				DocID: 1,
				Doc: `{
					"age": 34
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits(fieldId: "1") {
							delta
							docID
						}
					}
				`,
				Results: []map[string]any{
					{
						"delta": encrypt(testUtils.CBORValue(22)),
						"docID": "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
					},
					{
						"delta": encrypt(testUtils.CBORValue(21)),
						"docID": "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
					},
					{
						"delta": testUtils.CBORValue(34),
						"docID": "bae-d55bd956-1cc4-5d26-aa71-b98807ad49d6",
					},
					{
						"delta": testUtils.CBORValue(33),
						"docID": "bae-d55bd956-1cc4-5d26-aa71-b98807ad49d6",
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
