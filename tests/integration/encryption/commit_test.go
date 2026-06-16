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

package encryption

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestDocEncryption_WithEncryptionOnLWWCRDT_ShouldStoreCommitsDeltaEncrypted(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()
	nameCid := testUtils.NewSameValue()
	ageCid := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			addUserCollection(),
			&action.AddDoc{
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			&action.Request{
				Request: `
					query {
						_commits {
							cid
							delta
							docID
							fieldName
							height
							links {
								cid
								fieldName
							}
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":       gomega.And(ageCid, uniqueCid),
							"delta":     encryptedCBORValueWithKey(testUtils.CBORValue(21), shortDocIDKey(1), ""),
							"docID":     testUtils.NewDocIndex(0, 0),
							"fieldName": "age",
							"height":    int64(1),
							"links":     []map[string]any{},
						},
						{
							"cid":       gomega.And(nameCid, uniqueCid),
							"delta":     encryptedCBORValueWithKey(testUtils.CBORValue("John"), shortDocIDKey(1), ""),
							"docID":     testUtils.NewDocIndex(0, 0),
							"fieldName": "name",
							"height":    int64(1),
							"links":     []map[string]any{},
						},
						{
							"cid":       uniqueCid,
							"delta":     nil,
							"docID":     testUtils.NewDocIndex(0, 0),
							"fieldName": "_C",
							"height":    int64(1),
							"links": []map[string]any{
								{
									"cid":       nameCid,
									"fieldName": "name",
								},
								{
									"cid":       ageCid,
									"fieldName": "age",
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_UponUpdateOnLWWCRDT_ShouldEncryptCommitDelta(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			addUserCollection(),
			&action.AddDoc{
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			&action.UpdateDoc{
				Doc: `{
					"age":	22
				}`,
			},
			&action.Request{
				Request: `
					query {
						_commits(filter: {fieldName: {_eq: "age"}}) {
							delta
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"delta": encryptedCBORValueWithKey(testUtils.CBORValue(22), shortDocIDKey(1), ""),
						},
						{
							"delta": encryptedCBORValueWithKey(testUtils.CBORValue(21), shortDocIDKey(1), ""),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_WithMultipleDocsUponUpdate_ShouldEncryptOnlyRelevantDocs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			addUserCollection(),
			&action.AddDoc{
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			&action.AddDoc{
				Doc: islam33Doc,
			},
			&action.UpdateDoc{
				DocID: 0,
				Doc: `{
					"age": 22
				}`,
			},
			&action.UpdateDoc{
				DocID: 1,
				Doc: `{
					"age": 34
				}`,
			},
			&action.Request{
				Request: `
					query {
						_commits(filter: {fieldName: {_eq: "age"}}) {
							delta
							docID
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"delta": testUtils.CBORValue(34),
							"docID": testUtils.NewDocIndex(0, 1),
						},
						{
							"delta": testUtils.CBORValue(33),
							"docID": testUtils.NewDocIndex(0, 1),
						},
						{
							"delta": encryptedCBORValueWithKey(testUtils.CBORValue(22), shortDocIDKey(1), ""),
							"docID": testUtils.NewDocIndex(0, 0),
						},
						{
							"delta": encryptedCBORValueWithKey(testUtils.CBORValue(21), shortDocIDKey(1), ""),
							"docID": testUtils.NewDocIndex(0, 0),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_WithEncryptionOnCounterCRDT_ShouldStoreCommitsDeltaEncrypted(t *testing.T) {
	test := testUtils.TestCase{
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddCollection{
				SDL: `
                    type Users {
                        points: Int @crdt(type: pcounter)
                    }
                `},
			&action.AddDoc{
				Doc:            `{ "points": 5 }`,
				IsDocEncrypted: true,
			},
			&action.Request{
				Request: `
					query {
						_commits {
							delta
							docID
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"delta": encryptedCBORValueWithKey(testUtils.CBORValue(5), shortDocIDKey(1), ""),
							"docID": testUtils.NewDocIndex(0, 0),
						},
						{
							"delta": nil,
							"docID": testUtils.NewDocIndex(0, 0),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_UponUpdateOnCounterCRDT_ShouldEncryptedCommitDelta(t *testing.T) {
	test := testUtils.TestCase{
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddCollection{
				SDL: `
                    type Users {
                        points: Int @crdt(type: pcounter)
                    }
                `},
			&action.AddDoc{
				Doc:            `{ "points": 5 }`,
				IsDocEncrypted: true,
			},
			&action.UpdateDoc{
				Doc: `{
					"points": 3
				}`,
			},
			&action.Request{
				Request: `
					query {
						_commits(filter: {fieldName: {_eq: "points"}}) {
							delta
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"delta": encryptedCBORValueWithKey(testUtils.CBORValue(3), shortDocIDKey(1), ""),
						},
						{
							"delta": encryptedCBORValueWithKey(testUtils.CBORValue(5), shortDocIDKey(1), ""),
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
			addUserCollection(),
			&action.AddDoc{
				Doc:            "[" + islam33Doc + ", " + john21Doc + "]",
				IsDocEncrypted: true,
			},
			&action.Request{
				Request: `
					query {
						_commits {
							delta
							docID
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"delta": encryptedCBORValueWithKey(testUtils.CBORValue(33), shortDocIDKey(1), ""),
							"docID": testUtils.NewDocIndex(0, 0),
						},
						{
							"delta": encryptedCBORValueWithKey(testUtils.CBORValue("Islam"), shortDocIDKey(1), ""),
							"docID": testUtils.NewDocIndex(0, 0),
						},
						{
							"delta": nil,
							"docID": testUtils.NewDocIndex(0, 0),
						},
						{
							"delta": encryptedCBORValueWithKey(testUtils.CBORValue(21), shortDocIDKey(2), ""),
							"docID": testUtils.NewDocIndex(0, 1),
						},
						{
							"delta": encryptedCBORValueWithKey(testUtils.CBORValue("John"), shortDocIDKey(2), ""),
							"docID": testUtils.NewDocIndex(0, 1),
						},
						{
							"delta": nil,
							"docID": testUtils.NewDocIndex(0, 1),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_IfTwoDocsHaveSameFieldValue_CipherTextShouldBeDifferent(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			addUserCollection(),
			&action.AddDoc{
				Doc: `{
						"name": "John",
						"age": 21
					}`,
				IsDocEncrypted: true,
			},
			&action.AddDoc{
				Doc: `{
						"name": "Islam",
						"age": 21
					}`,
				IsDocEncrypted: true,
			},
			&action.Request{
				Request: `
					query {
						_commits(filter: {fieldName: {_eq: "age"}}) {
							delta
							fieldName
						}
					}
				`,
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					commits := testUtils.ConvertToArrayOfMaps(t, result["_commits"])
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
