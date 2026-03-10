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

package signature

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestDocSignature_WithEnabledSigning_ShouldQuery(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		Actions: []any{
			&action.AddCollection{
				SDL: `
                    type Users {
                        name: String
                        age: Int
                    }
                `},
			&action.AddDoc{
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
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

func TestDocSignature_WithEnabledSigning_ShouldQueryCommitsWithoutSignature(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		Actions: []any{
			&action.AddCollection{
				SDL: `
                    type Users {
                        name: String
                        age: Int
                    }
                `},
			&action.AddDoc{
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			&action.Request{
				Request: `
                    query {
                        Users {
                            _docID
                            name
                            age

							_version {
								height
							}
                        }
                    }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": testUtils.NewDocIndex(0, 0),
							"name":   "John",
							"age":    int64(21),
							"_version": []map[string]any{
								{
									"height": 1,
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
