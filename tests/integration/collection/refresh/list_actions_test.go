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

package refresh

import (
	"testing"
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	"github.com/sourcenetwork/defradb/tests/gen"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

func TestRefreshCollectionListAction_ListsUncompletedRefresh(t *testing.T) {
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{
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
			&action.Wait{
				// AddView refreshes on create, adding this wait pulls the events from *that* refresh, so that the
				// wait later in this test can pick up the only ones we care about.
				Action: immutable.Some(client.ActionExecution{
					CollectionID: "bafyreiex5xc3do2a42ymbt7tscfx7ifsdmgwkty6dbnfnktyiplqgeap4q",
					Action:       client.RefreshDatastoreAction,
					Status:       client.CompletedActionStatus,
				}),
				Duration: immutable.Some(time.Second),
			},
			testUtils.GenerateDocs{
				Options: []gen.Option{
					// Create enough users to allow the `ListActions` action to complete before the RefreshViews.
					gen.WithTypeDemand("Users", 100),
				},
				ForCollections: []string{"Users"},
			},
			&action.Async{
				Child: &action.ActionGroup{
					Children: []action.Action{
						&action.Wait{
							// Delay the start of the `ListActions` action until the view refresh has registered the action.
							Action: immutable.Some(client.ActionExecution{
								CollectionID: "bafyreiex5xc3do2a42ymbt7tscfx7ifsdmgwkty6dbnfnktyiplqgeap4q",
								Action:       client.RefreshDatastoreAction,
								Status:       client.InProgressActionStatus,
							}),
							Duration: immutable.Some(time.Second),
						},
						&action.ListActions{
							ExpectedInfo: []client.ActionExecution{
								{
									CollectionID: "bafyreiex5xc3do2a42ymbt7tscfx7ifsdmgwkty6dbnfnktyiplqgeap4q",
									Action:       client.RefreshDatastoreAction,
									Status:       client.InProgressActionStatus,
								},
							},
						},
					},
				},
			},
			&action.RefreshViews{},
			&action.Await{},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestRefreshCollectionListAction_DoesNotListCompletedRefreshes(t *testing.T) {
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
			// Execute the view refresh synchronously to ensure it has been completed
			// before the ListActions action executes.
			&action.RefreshViews{},
			&action.ListActions{
				ExpectedInfo: []client.ActionExecution{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
