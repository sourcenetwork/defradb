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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryption_WithEncryptionOnLWWCRDT_ShouldStoreCommitsDeltaEncrypted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							cid
							delta
							docID
							fieldName
							height
							links {
								cid
								name
							}
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"cid":       "bafyreia37txi77ajmma3t3o4hlnkx7qdbzymioplbuscla576i52rr5hri",
							"delta":     encrypt(testUtils.CBORValue(21), john21DocID, ""),
							"docID":     john21DocID,
							"fieldName": "age",
							"height":    int64(1),
							"links":     []map[string]any{},
						},
						{
							"cid":       "bafyreiajo5esphlst5bi2qjudn7uutk5layfa3edxb55rett54qj7gznai",
							"delta":     encrypt(testUtils.CBORValue("John"), john21DocID, ""),
							"docID":     john21DocID,
							"fieldName": "name",
							"height":    int64(1),
							"links":     []map[string]any{},
						},
						{
							"cid":       "bafyreicmnl5hzhq4q533a47igftavebkqjhxl22t3hag6yods5j6iydji4",
							"delta":     nil,
							"docID":     john21DocID,
							"fieldName": "_C",
							"height":    int64(1),
							"links": []map[string]any{
								{
									"cid":  "bafyreia37txi77ajmma3t3o4hlnkx7qdbzymioplbuscla576i52rr5hri",
									"name": "age",
								},
								{
									"cid":  "bafyreiajo5esphlst5bi2qjudn7uutk5layfa3edxb55rett54qj7gznai",
									"name": "name",
								},
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
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits(fieldName: "age") {
							delta
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"delta": encrypt(testUtils.CBORValue(22), john21DocID, ""),
						},
						{
							"delta": encrypt(testUtils.CBORValue(21), john21DocID, ""),
						},
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
				Doc:            john21Doc,
				IsDocEncrypted: true,
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
						commits(fieldName: "age") {
							delta
							docID
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"delta": testUtils.CBORValue(34),
							"docID": islam33DocID,
						},
						{
							"delta": testUtils.CBORValue(33),
							"docID": islam33DocID,
						},
						{
							"delta": encrypt(testUtils.CBORValue(22), john21DocID, ""),
							"docID": john21DocID,
						},
						{
							"delta": encrypt(testUtils.CBORValue(21), john21DocID, ""),
							"docID": john21DocID,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_WithEncryptionOnCounterCRDT_ShouldStoreCommitsDeltaEncrypted(t *testing.T) {
	const docID = "bae-b0875c46-79f8-568f-9cc9-13eab3e2b8b1"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
                    type Users {
                        points: Int @crdt(type: pcounter)
                    }
                `},
			testUtils.CreateDoc{
				Doc:            `{ "points": 5 }`,
				IsDocEncrypted: true,
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
				Results: map[string]any{
					"commits": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_UponUpdateOnCounterCRDT_ShouldEncryptedCommitDelta(t *testing.T) {
	const docID = "bae-b0875c46-79f8-568f-9cc9-13eab3e2b8b1"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
                    type Users {
                        points: Int @crdt(type: pcounter)
                    }
                `},
			testUtils.CreateDoc{
				Doc:            `{ "points": 5 }`,
				IsDocEncrypted: true,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"points": 3
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						commits(fieldName: "points") {
							delta
						}
					}
				`,
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"delta": encrypt(testUtils.CBORValue(3), docID, ""),
						},
						{
							"delta": encrypt(testUtils.CBORValue(5), docID, ""),
						},
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
				Doc:            "[" + islam33Doc + ", " + john21Doc + "]",
				IsDocEncrypted: true,
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
				Results: map[string]any{
					"commits": []map[string]any{
						{
							"delta": encrypt(testUtils.CBORValue(33), islam33DocID, ""),
							"docID": testUtils.NewDocIndex(0, 0),
						},
						{
							"delta": encrypt(testUtils.CBORValue("Islam"), islam33DocID, ""),
							"docID": testUtils.NewDocIndex(0, 0),
						},
						{
							"delta": nil,
							"docID": testUtils.NewDocIndex(0, 0),
						},
						{
							"delta": encrypt(testUtils.CBORValue(21), john21DocID, ""),
							"docID": testUtils.NewDocIndex(0, 1),
						},
						{
							"delta": encrypt(testUtils.CBORValue("John"), john21DocID, ""),
							"docID": testUtils.NewDocIndex(0, 1),
						},
						{
							"delta": nil,
							"docID": testUtils.NewDocIndex(0, 1),
						},
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
				IsDocEncrypted: true,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name": "Islam",
						"age": 21
					}`,
				IsDocEncrypted: true,
			},
			testUtils.Request{
				Request: `
					query {
						commits(fieldName: "age") {
							delta
							fieldName
						}
					}
				`,
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					commits := testUtils.ConvertToArrayOfMaps(t, result["commits"])
					require.Equal(t, 2, len(commits), "Expected 2 commits")
					require.Equal(t, commits[0]["fieldName"], "age")
					delta1 := commits[0]["delta"]
					delta2 := commits[1]["delta"]
					assert.NotEqual(t, delta1, delta2, "docs should be encrypted with different encryption keys")
					return true, ""
				}),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
