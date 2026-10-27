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

package simple

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithDocIDFilterBlock(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {_docID: {_eq: "{{.DocID0_0}}"}}) {
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(21),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithDocIDFilterInBlock(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.Request{
				Request: `query {
					Users(filter: {_docID: {_in: ["{{.DocID0_0}}", "{{.DocID0_2}}"]}}) {
						_docID
						Name
						Age
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "{{.DocID0_0}}",
							"Name":   "John",
							"Age":    int64(21),
						},
						{
							"_docID": "{{.DocID0_2}}",
							"Name":   "Alice",
							"Age":    int64(40),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleOrdersByDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 40
				}`,
			},
			&action.Request{
				Request: `query {
					Users(order: {_docID: ASC}) {
						_docID
						Name
						Age
					}
				}`,
				Asserter: testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
					docs := action.ConvertToArrayOfMaps(t, result["Users"])
					require.Len(t, docs, 3)

					docIDs := make([]string, len(docs))
					names := make([]string, len(docs))
					for i, doc := range docs {
						docID, ok := doc["_docID"].(string)
						require.True(t, ok)
						docIDs[i] = docID

						name, ok := doc["Name"].(string)
						require.True(t, ok)
						names[i] = name
					}

					require.True(t, slices.IsSorted(docIDs), "doc IDs should be sorted ascending: %v", docIDs)
					require.ElementsMatch(t, []string{"John", "Bob", "Alice"}, names)
					return true, ""
				}),
			},
		},
	}

	executeTestCase(t, test)
}
