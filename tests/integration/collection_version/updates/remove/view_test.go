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

package remove

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateRemoveView(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddView{
				SDL: `
					type UserView @materialized(if: false) {
						name: String
					}
				`,
				Query: `
					Users {
						name
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/UserView"
						}
					]
				`,
			},
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetCollectionName("UserView"),
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateRemoveNonMaterializedViewWithData(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddView{
				SDL: `
					type UserView @materialized(if: false) {
						name: String
					}
				`,
				Query: `
					Users {
						name
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/UserView"
						}
					]
				`,
			},
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetCollectionName("UserView"),
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateRemoveMaterializedViewWithUnrefreshedData(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddView{
				SDL: `
					type UserView @materialized(if: true) {
						name: String
					}
				`,
				Query: `
					Users {
						name
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.PatchCollection{
				// We are removing the view *before* the view has been refreshed, it should be deleted
				// as there is no reason for us to not be able to delete empty datasets - there are no
				// complications such as secondary indexes.
				Patch: `
					[
						{
							"op": "remove",
							"path": "/UserView"
						}
					]
				`,
			},
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetCollectionName("UserView"),
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateRemoveMaterializedViewWithRefreshedData(t *testing.T) {
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{
			// The expected error should only occur when using a materialized view.
			testUtils.MaterializedViewType,
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
				SDL: `
					type UserView @materialized(if: true) {
						name: String
					}
				`,
				Query: `
					Users {
						name
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.RefreshViews{},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/UserView"
						}
					]
				`,
				ExpectedError: "cannot delete a collection that has documents, first delete the documents and then delete the version",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateRemoveCollectionBackingUnmaterializedView(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddView{
				SDL: `
					type UserView @materialized(if: false) {
						name: String
					}
				`,
				Query: `
					Users {
						name
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/Users"
						}
					]
				`,
			},
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetCollectionName("Users"),
				ExpectedResults: []client.CollectionVersion{},
			},
			&action.Request{
				Request: `query {
					UserView {
						name
					}
				}`,
				ExpectedError: `collection not found`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateRemoveCollectionBackingMaterializedView(t *testing.T) {
	test := testUtils.TestCase{
		// The view multiplier currently refreshes views as part of the `Request`
		// action - this changes the test definition in a way that we do not want here.
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddView{
				SDL: `
					type UserView @materialized(if: true) {
						name: String
					}
				`,
				Query: `
					Users {
						name
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/Users"
						}
					]
				`,
			},
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetCollectionName("Users"),
				ExpectedResults: []client.CollectionVersion{},
			},
			&action.Request{
				Request: `query {
					UserView {
						name
					}
				}`,
				Results: map[string]any{
					"UserView": []map[string]any{},
				},
			},
			&action.RefreshViews{
				ExpectedError: "key not found",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
