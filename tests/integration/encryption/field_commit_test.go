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

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocEncryptionField_WithEncryptionOnField_ShouldStoreOnlyFieldsDeltaEncrypted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			addUserCollection(),
			&action.AddDoc{
				Doc:             john21Doc,
				EncryptedFields: []string{"age"},
			},
			&action.Request{
				Request: `
					query {
						_commits {
							delta
							docID
							fieldName
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
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
							"fieldName": "_C",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryptionField_WithDocAndFieldEncryption_ShouldUseDedicatedEncKeyForIndividualFields(t *testing.T) {
	deltaForField := func(fieldName string, result []map[string]any) any {
		for _, r := range result {
			if r["fieldName"] == fieldName {
				return r["delta"]
			}
		}
		t.Fatalf("Field %s not found in results %v", fieldName, result)
		return nil
	}

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name1: String
						name2: String
						name3: String
						name4: String
					}`,
			},
			&action.AddDoc{
				Doc: `{
						"name1": "John",
						"name2": "John",
						"name3": "John",
						"name4": "John"
					}`,
				IsDocEncrypted:  true,
				EncryptedFields: []string{"name1", "name3"},
			},
			&action.Request{
				Request: `
					query {
						_commits {
							cid
							delta
							fieldName
						}
					}
				`,
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					commits := testUtils.ConvertToArrayOfMaps(t, result["_commits"])
					name1 := deltaForField("name1", commits)
					name2 := deltaForField("name2", commits)
					name3 := deltaForField("name3", commits)
					name4 := deltaForField("name4", commits)
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

func TestDocEncryptionField_UponUpdateWithDocAndFieldEncryption_ShouldUseDedicatedEncKeyForIndividualFields(t *testing.T) {
	deltaForField := func(fieldName string, result []map[string]any) any {
		for _, r := range result {
			if r["fieldName"] == fieldName {
				return r["delta"]
			}
		}
		t.Fatalf("Field %s not found in results %v", fieldName, result)
		return nil
	}

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name1: String
						name2: String
						name3: String
						name4: String
					}`,
			},
			&action.AddDoc{
				Doc: `{
						"name1": "John",
						"name2": "John",
						"name3": "John",
						"name4": "John"
					}`,
				IsDocEncrypted:  true,
				EncryptedFields: []string{"name1", "name3"},
			},
			&action.UpdateDoc{
				Doc: `{
					"name1": "Andy",
					"name2": "Andy",
					"name3": "Andy",
					"name4": "Andy"
				}`,
			},
			&action.Request{
				Request: `
					query {
						_commits(order: {height: DESC}, limit: 5) {
							cid
							delta
							fieldName
							height
						}
					}
				`,
				Asserter: testUtils.ResultAsserterFunc(func(_ testing.TB, result map[string]any) (bool, string) {
					commits := testUtils.ConvertToArrayOfMaps(t, result["_commits"])
					name1 := deltaForField("name1", commits)
					name2 := deltaForField("name2", commits)
					name3 := deltaForField("name3", commits)
					name4 := deltaForField("name4", commits)
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
