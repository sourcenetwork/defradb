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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocEncryption_WithEncryptionOnLWWCRDT_ShouldStoreCommitsDeltaEncrypted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc:         john21Doc,
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
						"cid":          "bafyreibdjepzhhiez4o27srv33xcd52yr336tpzqtkv36rdf3h3oue2l5m",
						"collectionID": int64(1),
						"delta":        encrypt(testUtils.CBORValue(21), john21DocID, ""),
						"docID":        john21DocID,
						"fieldId":      "1",
						"fieldName":    "age",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreihkiua7jpwkye3xlex6s5hh2azckcaljfi2h3iscgub5sikacyrbu",
						"collectionID": int64(1),
						"delta":        encrypt(testUtils.CBORValue("John"), john21DocID, ""),
						"docID":        john21DocID,
						"fieldId":      "2",
						"fieldName":    "name",
						"height":       int64(1),
						"links":        []map[string]any{},
					},
					{
						"cid":          "bafyreidxdhzhwjrv5s4x6cho5drz6xq2tc7oymzupf4p4gfk6eelsnc7ke",
						"collectionID": int64(1),
						"delta":        nil,
						"docID":        john21DocID,
						"fieldId":      "C",
						"fieldName":    nil,
						"height":       int64(1),
						"links": []map[string]any{
							{
								"cid":  "bafyreibdjepzhhiez4o27srv33xcd52yr336tpzqtkv36rdf3h3oue2l5m",
								"name": "age",
							},
							{
								"cid":  "bafyreihkiua7jpwkye3xlex6s5hh2azckcaljfi2h3iscgub5sikacyrbu",
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

func TestDocEncryption_UponUpdateOnLWWCRDT_ShouldEncryptCommitDelta(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc:         john21Doc,
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
						"delta": encrypt(testUtils.CBORValue(22), john21DocID, ""),
					},
					{
						"delta": encrypt(testUtils.CBORValue(21), john21DocID, ""),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_WithMultipleDocsUponUpdate_ShouldEncryptOnlyRelevantDocs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc:         john21Doc,
				IsEncrypted: true,
			},
			testUtils.CreateDoc{
				Doc: islam33Doc,
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
						"delta": encrypt(testUtils.CBORValue(22), john21DocID, ""),
						"docID": john21DocID,
					},
					{
						"delta": encrypt(testUtils.CBORValue(21), john21DocID, ""),
						"docID": john21DocID,
					},
					{
						"delta": testUtils.CBORValue(34),
						"docID": islam33DocID,
					},
					{
						"delta": testUtils.CBORValue(33),
						"docID": islam33DocID,
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
							delta
							docID
						}
					}
				`,
				Results: []map[string]any{
					{
						"delta": encrypt(testUtils.CBORValue(5), docID, ""),
						"docID": docID,
					},
					{
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
						"delta": encrypt(testUtils.CBORValue(3), docID, ""),
					},
					{
						"delta": encrypt(testUtils.CBORValue(5), docID, ""),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_UponEncryptionSeveralDocs_ShouldStoreAllCommitsDeltaEncrypted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc:         "[" + john21Doc + ", " + islam33Doc + "]",
				IsEncrypted: true,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							delta
							docID
						}
					}
				`,
				Results: []map[string]any{
					{
						"delta": encrypt(testUtils.CBORValue(21), john21DocID, ""),
						"docID": testUtils.NewDocIndex(0, 0),
					},
					{
						"delta": encrypt(testUtils.CBORValue("John"), john21DocID, ""),
						"docID": testUtils.NewDocIndex(0, 0),
					},
					{
						"delta": nil,
						"docID": testUtils.NewDocIndex(0, 0),
					},
					{
						"delta": encrypt(testUtils.CBORValue(33), islam33DocID, ""),
						"docID": testUtils.NewDocIndex(0, 1),
					},
					{
						"delta": encrypt(testUtils.CBORValue("Islam"), islam33DocID, ""),
						"docID": testUtils.NewDocIndex(0, 1),
					},
					{
						"delta": nil,
						"docID": testUtils.NewDocIndex(0, 1),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_IfTwoDocsHaveSameFieldValue_CipherTextShouldBeDifferent(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name": "John",
						"age": 21
					}`,
				IsEncrypted: true,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name": "Islam",
						"age": 21
					}`,
				IsEncrypted: true,
			},
			testUtils.Request{
				Request: `
					query {
						commits(fieldId: "1") {
							delta
							fieldName
						}
					}
				`,
				Asserter: testUtils.ResultAsserterFunc(func(_ testing.TB, result []map[string]any) (bool, string) {
					require.Equal(t, 2, len(result), "Expected 2 commits")
					require.Equal(t, result[0]["fieldName"], "age")
					delta1 := result[0]["delta"].([]byte)
					delta2 := result[1]["delta"].([]byte)
					assert.NotEqual(t, delta1, delta2, "docs should be encrypted with different encryption keys")
					return true, ""
				}),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
