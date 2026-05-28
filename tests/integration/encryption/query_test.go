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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestDocEncryption_WithEncryption_ShouldFetchDecrypted(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
                    type Users {
                        name: String
                        age: Int
                    }
                `},
			&action.AddDoc{
				Doc:            john21Doc,
				IsDocEncrypted: true,
			},
			&action.Request{
				Request: `
                    query {
                        Users {
                            _docID
                            name
                            age
                        }
                    }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": testUtils.NewDocIndex(0, 0),
							"name":   "John",
							"age":    int64(21),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDocEncryption_WithEncryptionOnCounterCRDT_ShouldFetchDecrypted(t *testing.T) {
	const query = `
		query {
			Users {
				name
				points
			}
		}`

	test := testUtils.TestCase{
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddCollection{
				SDL: `
                    type Users {
                        name: String
                        points: Int @crdt(type: pcounter)
                    }
                `},
			&action.AddDoc{
				Doc: `{
						"name":	"John",
						"points": 5
					}`,
				IsDocEncrypted: true,
			},
			&action.Request{
				Request: query,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"points": 5,
						},
					},
				},
			},
			&action.UpdateDoc{
				DocID: 0,
				Doc:   `{ "points": 3 }`,
			},
			&action.Request{
				Request: query,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"points": 8,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
