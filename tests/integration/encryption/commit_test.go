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
						"cid":          "bafyreih7ry7ef26xn3lm2rhxusf2rbgyvl535tltrt6ehpwtvdnhlmptiu",
						"collectionID": int64(1),
						"delta":        encrypt(testUtils.CBORValue(21)),
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreifusejlwidaqswasct37eorazlfix6vyyn5af42pmjvktilzj5cty",
						"collectionID": int64(1),
						"delta":        encrypt(testUtils.CBORValue("John")),
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "2",
						"fieldName":    "name",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreicvxlfxeqghmc3gy56rp5rzfejnbng4nu77x5e3wjinfydl6wvycq",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3",
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(1),
						"links": []map[string]any{
							{
								"cid":  "bafyreifusejlwidaqswasct37eorazlfix6vyyn5af42pmjvktilzj5cty",
								"name": "name",
							},
							{
								"cid":  "bafyreih7ry7ef26xn3lm2rhxusf2rbgyvl535tltrt6ehpwtvdnhlmptiu",
								"name": "age",
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
