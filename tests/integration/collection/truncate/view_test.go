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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestTruncateCollectionViewAdd_RemovesDocument(t *testing.T) {
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{
			state.MaterializedViewType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddView{
				Query: `
					Users {
						name
					}
				`,
				SDL: `
					type UserView {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "Fred",
				},
			},
			&action.RefreshViews{},
			&action.Truncate{
				// Truncate the View, but not the underlying collection
				CollectionIndex: 1,
			},
			&action.Request{
				DoNotRefreshViews: true,
				Request: `query {
					UserView {
						name
					}
				}`,
				Results: map[string]any{
					"UserView": []map[string]any{},
				},
			},
			// Refresh the View and assert that it has been reconstructed from the underlying
			// collection.
			&action.RefreshViews{},
			&action.Request{
				DoNotRefreshViews: true,
				NonOrderedResults: true,
				Request: `query {
					UserView {
						name
					}
				}`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"name": "John",
						},
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
func TestTruncateCollectionViewAdd_TruncatingSourceDoesNotTruncateView(t *testing.T) {
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{
			state.MaterializedViewType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddView{
				Query: `
					Users {
						name
					}
				`,
				SDL: `
					type UserView {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "Fred",
				},
			},
			&action.RefreshViews{},
			&action.Truncate{
				// Truncate the underlying collection, without truncating the view
				CollectionIndex: 0,
			},
			&action.Request{
				DoNotRefreshViews: true,
				NonOrderedResults: true,
				Request: `query {
					UserView {
						name
					}
				}`,
				Results: map[string]any{
					"UserView": []map[string]any{
						{
							"name": "John",
						},
						{
							"name": "Fred",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
