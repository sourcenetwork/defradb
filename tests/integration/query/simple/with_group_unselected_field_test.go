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
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This is a regression test for https://github.com/sourcenetwork/defradb/issues/4954.
//
// Grouping must still occur on the grouped-by field even when that field is not part of
// the parent's selection set. Previously the field was never fetched, so every document
// produced an identical (nil) group key and all documents collapsed into a single group.
func TestQuerySimpleWithGroupByNumberWithoutRenderedGroupField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 32
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
					"Name": "Carlo",
					"Age": 55
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Alice",
					"Age": 19
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Age]) {
						GROUP {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"GROUP": []map[string]any{
								{"Name": "John"},
								{"Name": "Bob"},
							},
						},
						{
							"GROUP": []map[string]any{
								{"Name": "Carlo"},
							},
						},
						{
							"GROUP": []map[string]any{
								{"Name": "Alice"},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

// Companion to the above using a String grouped-by field (matching the genre field in the
// original issue report), ensuring the fix is not specific to a single field kind.
func TestQuerySimpleWithGroupByStringWithoutRenderedGroupField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Email": "fiction@example.com"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Email": "fiction@example.com"
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Carlo",
					"Email": "nonfiction@example.com"
				}`,
			},
			&action.Request{
				Request: `query {
					Users(groupBy: [Email]) {
						GROUP {
							Name
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"GROUP": []map[string]any{
								{"Name": "John"},
								{"Name": "Bob"},
							},
						},
						{
							"GROUP": []map[string]any{
								{"Name": "Carlo"},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}
