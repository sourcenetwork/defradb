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

package truncate

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestTruncateCollectionIndexFilter_RemovesDocument(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String @index
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Truncate{
				CollectionIndex: 0,
			},
			&action.Request{
				Request: `query {
					Users(filter: {name: {_eq: "John"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTruncateCollectionIndexFilter_WithUniqueIndex_RemovesDocument(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String @index(unique: true)
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Truncate{
				CollectionIndex: 0,
			},
			&action.Request{
				Request: `query {
					Users(filter: {name: {_eq: "John"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTruncateCollectionIndexFilter_WithUniqueIndex_AllowsRecreationOfDocument(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String @index(unique: true)
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Truncate{
				CollectionIndex: 0,
			},
			&action.AddDoc{
				CollectionID: 0,
				// If the unique index had not been deleted, then this create would
				// error, as the unique index would have been violated.
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Request{
				Request: `query {
					Users(filter: {name: {_eq: "John"}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
