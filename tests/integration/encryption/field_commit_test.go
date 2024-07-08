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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryptionField_WithEncryptionOnField_ShouldStoreOnlyFieldsDeltaEncrypted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc:             john21Doc,
				EncryptedFields: []string{"age"},
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							delta
							docID
							fieldName
						}
					}
				`,
				Results: []map[string]any{
					{
						"delta":     encrypt(testUtils.CBORValue(21), john21DocID, "age"),
						"docID":     john21DocID,
						"fieldName": "age",
					},
					{
						"delta":     testUtils.CBORValue("John"),
						"docID":     john21DocID,
						"fieldName": "name",
					},
					{
						"delta":     nil,
						"docID":     john21DocID,
						"fieldName": nil,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionField_WithDocAndFieldEncryption_ShouldUseDedicatedEncKeyForIndividualFields(t *testing.T) {
	deltaForField := func(fieldName string, result []map[string]any) []byte {
		for _, r := range result {
			if r["fieldName"] == fieldName {
				return r["delta"].([]byte)
			}
		}
		t.Fatalf("Field %s not found in results %v", fieldName, result)
		return nil
	}

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name1: String
						name2: String
						name3: String
						name4: String
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name1": "John",
						"name2": "John",
						"name3": "John",
						"name4": "John"
					}`,
				IsEncrypted:     true,
				EncryptedFields: []string{"name1", "name3"},
			},
			testUtils.Request{
				Request: `
					query {
						commits {
							cid
							delta
							fieldName
						}
					}
				`,
				Asserter: testUtils.ResultAsserterFunc(func(_ testing.TB, result []map[string]any) (bool, string) {
					name1 := deltaForField("name1", result)
					name2 := deltaForField("name2", result)
					name3 := deltaForField("name3", result)
					name4 := deltaForField("name4", result)
					assert.Equal(t, name2, name4, "name2 and name4 should have the same encryption key")
					assert.NotEqual(t, name2, name1, "name2 and name1 should have different encryption keys")
					assert.NotEqual(t, name2, name3, "name2 and name3 should have different encryption keys")
					assert.NotEqual(t, name1, name3, "name1 and name3 should have different encryption keys")
					return true, ""
				}),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionField_WithDocAndFieldEncryption_ShouldUseDedicatedEncKeyForIndividualFields2(t *testing.T) {
	deltaForField := func(fieldName string, result []map[string]any) []byte {
		for _, r := range result {
			if r["fieldName"] == fieldName {
				return r["delta"].([]byte)
			}
		}
		t.Fatalf("Field %s not found in results %v", fieldName, result)
		return nil
	}

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name1: String
						name2: String
						name3: String
						name4: String
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"name1": "John",
						"name2": "John",
						"name3": "John",
						"name4": "John"
					}`,
				IsEncrypted:     true,
				EncryptedFields: []string{"name1", "name3"},
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name1": "Andy",
					"name2": "Andy",
					"name3": "Andy",
					"name4": "Andy"
				}`,
			},
			testUtils.Request{
				Request: `
					query {

						commits(order: {height: DESC}, limit: 5) {
							cid
							delta
							fieldName
							height
						}
					}
				`,
				Asserter: testUtils.ResultAsserterFunc(func(_ testing.TB, result []map[string]any) (bool, string) {
					name1 := deltaForField("name1", result)
					name2 := deltaForField("name2", result)
					name3 := deltaForField("name3", result)
					name4 := deltaForField("name4", result)
					assert.Equal(t, name2, name4, "name2 and name4 should have the same encryption key")
					assert.NotEqual(t, name2, name1, "name2 and name1 should have different encryption keys")
					assert.NotEqual(t, name2, name3, "name2 and name3 should have different encryption keys")
					assert.NotEqual(t, name1, name3, "name1 and name3 should have different encryption keys")
					return true, ""
				}),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
