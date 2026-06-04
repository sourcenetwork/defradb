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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestView_SimpleMaterialized_AutoUpdatesOnViewAdd(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.AddView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView {
						name: String
					}
				`,
			},
			&action.Request{
				// We are testing that the refresh occurs on view create, so we must disable
				// the test framework's auto-refresh done within this Request's execution in
				// order to test it.
				DoNotRefreshViews: true,
				Request: `query {
							UserView {
								name
							}
						}`,
				Results: map[string]any{
					// Even though UserView was added after the document was added, the results are
					// present because the view will automatically refresh upon its addition.
					"UserView": []map[string]any{
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

func TestView_SimpleMaterialized_RefreshesAfterEarlierRefresh(t *testing.T) {
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{
			testUtils.MaterializedViewType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.AddView{
				Query: `
					User {
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
				Doc: `{
					"name":	"Fred"
				}`,
			},
			// Refresh the view after an earlier refresh (with data).  We had a bug here
			// where RefreshViews would fail only if there was already data in the view cache.
			&action.RefreshViews{},
			&action.Request{
				// It doesn't really matter if it refreshes again, but it is a bit wasteful,
				// and it is nicer to be explicit for this test.
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

func TestView_SimpleMaterialized_DoesNotAutoUpdate(t *testing.T) {
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{
			testUtils.MaterializedViewType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name":	"John"
				}`,
			},
			&action.AddView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView {
						name: String
					}
				`,
			},
			&action.RefreshViews{},
			&action.AddDoc{
				Doc: `{
					"name":	"Fred"
				}`,
			},
			&action.Request{
				// Disable the test framework's auto-refreshing of views for this test
				// so that we may verify the behaviour when the views are not refreshed
				DoNotRefreshViews: true,
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
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
