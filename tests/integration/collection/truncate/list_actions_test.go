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
	"time"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	"github.com/sourcenetwork/defradb/tests/gen"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/immutable"
)

func TestTruncateCollectionListAction_ListsUncompletedTruncates(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.GenerateDocs{
				// Create enough users to allow the `ListActions` action to complete before the truncate.
				Options: []gen.Option{
					gen.WithTypeDemand("Users", 100),
				},
			},
			&action.Async{
				Child: &action.ActionGroup{
					Children: []action.Action{
						&action.Wait{
							// Delay the start of the `ListActions` action until the view refresh has registered the action.
							Action: immutable.Some(client.ActionExecution{
								CollectionID: "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
								Action:       client.RefreshDatastoreAction,
								Status:       client.InProgressActionStatus,
							}),
							Duration: immutable.Some(time.Second),
						},
						&action.ListActions{
							ExpectedInfo: []client.ActionExecution{
								{
									CollectionID: "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
									Action:       client.TruncateAction,
									Status:       client.InProgressActionStatus,
								},
							},
						},
					},
				},
			},
			&action.Truncate{},
			&action.Await{},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTruncateCollectionListAction_DoesNotListCompletedTruncates(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			// Execute the truncate synchronously to ensure it has been completed
			// before the ListActions action executes.
			&action.Truncate{},
			&action.ListActions{
				ExpectedInfo: []client.ActionExecution{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
