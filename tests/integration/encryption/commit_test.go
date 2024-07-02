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

func TestDocEncryption_WithEncryptionOnLWWCRDT_ShouldStoreCommitsDeltaEncrypted(t *testing.T) {
	const docID = "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3"

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
						"docID":        docID,
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreifusejlwidaqswasct37eorazlfix6vyyn5af42pmjvktilzj5cty",
						"collectionID": int64(1),
						"delta":        encrypt(testUtils.CBORValue("John")),
						"docID":        docID,
						"fieldId":      "2",
						"fieldName":    "name",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreicvxlfxeqghmc3gy56rp5rzfejnbng4nu77x5e3wjinfydl6wvycq",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        docID,
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

func TestDocEncryption_UponUpdateOnLWWCRDT_ShouldEncryptCommitDelta(t *testing.T) {
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

func TestDocEncryption_WithMultipleDocsUponUpdate_ShouldEncryptOnlyRelevantDocs(t *testing.T) {
	const johnDocID = "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3"
	const islamDocID = "bae-d55bd956-1cc4-5d26-aa71-b98807ad49d6"

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
			testUtils.CreateDoc{
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
						"docID": johnDocID,
					},
					{
						"delta": encrypt(testUtils.CBORValue(21)),
						"docID": johnDocID,
					},
					{
						"delta": testUtils.CBORValue(34),
						"docID": islamDocID,
					},
					{
						"delta": testUtils.CBORValue(33),
						"docID": islamDocID,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_WithEncryptionOnCounterCRDT_ShouldStoreCommitsDeltaEncrypted(t *testing.T) {
	const docID = "bae-d3cc98b4-38d5-5c50-85a3-d3045d44094e"

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
                    type Users {
                        points: Int @crdt(type: "pcounter")
                    }
                `},
			testUtils.CreateDoc{
				Doc:         `{ "points": 5 }`,
				IsEncrypted: true,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							cid
							delta
							docID
						}
					}
				`,
				Results: []map[string]any{
					{
						"cid":   "bafyreieb6owsoljj4vondkx35ngxmhliauwvphicz4edufcy7biexij7mu",
						"delta": encrypt(testUtils.CBORValue(5)),
						"docID": docID,
					},
					{
						"cid":   "bafyreif2lejhvdja2rmo237lrwpj45usrm55h6gzr4ewl6gajq3cl4ppsi",
						"delta": nil,
						"docID": docID,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_UponUpdateOnCounterCRDT_ShouldEncryptedCommitDelta(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
                    type Users {
                        points: Int @crdt(type: "pcounter")
                    }
                `},
			testUtils.CreateDoc{
				Doc:         `{ "points": 5 }`,
				IsEncrypted: true,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"points": 3
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
						"delta": encrypt(testUtils.CBORValue(3)),
					},
					{
						"delta": encrypt(testUtils.CBORValue(5)),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_UponEncryptionSeveralDocs_ShouldStoreAllCommitsDeltaEncrypted(t *testing.T) {
	const johnDocID = "bae-c9fb0fa4-1195-589c-aa54-e68333fb90b3"
	const islamDocID = "bae-d55bd956-1cc4-5d26-aa71-b98807ad49d6"

	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `[{
						"name":	"John",
						"age":	21
					},
					{
						"name":	"Islam",
						"age":	33
					}]`,
				IsEncrypted: true,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							cid
							delta
							docID
						}
					}
				`,
				Results: []map[string]any{
					{
						"cid":   "bafyreih7ry7ef26xn3lm2rhxusf2rbgyvl535tltrt6ehpwtvdnhlmptiu",
						"delta": encrypt(testUtils.CBORValue(21)),
						"docID": johnDocID,
					},
					{
						"cid":   "bafyreifusejlwidaqswasct37eorazlfix6vyyn5af42pmjvktilzj5cty",
						"delta": encrypt(testUtils.CBORValue("John")),
						"docID": johnDocID,
					},
					{
						"cid":   "bafyreicvxlfxeqghmc3gy56rp5rzfejnbng4nu77x5e3wjinfydl6wvycq",
						"delta": nil,
						"docID": johnDocID,
					},
					{
						"cid":   "bafyreibe24bo67owxewoso3ekinera2bhusguij5qy2ahgyufaq3fbvaxa",
						"delta": encrypt(testUtils.CBORValue(33)),
						"docID": islamDocID,
					},
					{
						"cid":   "bafyreie2fddpidgc62fhd2fjrsucq3spgh2mgvto2xwolcdmdhb5pdeok4",
						"delta": encrypt(testUtils.CBORValue("Islam")),
						"docID": islamDocID,
					},
					{
						"cid":   "bafyreifulxdkf4m3wmmdxjg43l4mw7uuxl5il27eabklc22nptilrh64sa",
						"delta": nil,
						"docID": islamDocID,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
