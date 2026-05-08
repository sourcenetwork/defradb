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

package delete

import (
	"testing"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
	"github.com/sourcenetwork/immutable"
)

func TestDeleteCollection_Simple_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.DeleteCollection{
				ActiveOnly: true,
				Names:      []string{"Users"},
			},
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteCollection_Simple_QueriesNoLongerWork(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.DeleteCollection{
				ActiveOnly: true,
				Names:      []string{"Users"},
			},
			&action.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				ExpectedError: `Cannot query field "Users" on type "Query".`,
			},
			&action.Request{
				Request: `mutation {
					add_Users(input:{}) {
						name
					}
				}`,
				ExpectedError: `Cannot query field "add_Users" on type "Mutation".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteCollection_WithDocuments_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.DeleteCollection{
				ActiveOnly:    true,
				Names:         []string{"Users"},
				ExpectedError: "cannot delete a collection that has documents",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteCollection_WithSoftDeletedDocuments_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.DeleteDoc{
				CollectionID: 0,
				DocID:        0,
			},
			&action.DeleteCollection{
				ActiveOnly:    true,
				Names:         []string{"Users"},
				ExpectedError: "cannot delete a collection that has documents",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteCollection_NonExistentCollection_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.DeleteCollection{
				ActiveOnly:    true,
				Names:         []string{"NonExistent"},
				ExpectedError: "collection not found",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteCollection_MultipleCollections_DeleteOne(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
					type Books {
						title: String
					}
				`,
			},
			&action.DeleteCollection{
				ActiveOnly: true,
				Names:      []string{"Users"},
			},
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Books",
						IsMaterialized: true,
						IsActive:       true,
					},
				},
			},
			&action.Request{
				Request: `query {
					Books {
						title
					}
				}`,
				Results: map[string]any{
					"Books": []map[string]any{},
				},
			},
			&action.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				ExpectedError: `Cannot query field "Users" on type "Query".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteCollection_WithTransaction_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.DeleteCollection{
				ActiveOnly:    true,
				TransactionID: immutable.Some(1),
				Names:         []string{"Users"},
			},
			&action.CommitTransaction{
				TransactionID: 1,
			},
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteCollection_WithTransactionWithoutCommit_CollectionStillExists(t *testing.T) {
	test := testUtils.TestCase{
		// LevelDB does not support concurrent transactions
		// todo: https://github.com/sourcenetwork/defradb/issues/4442
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			testUtils.BadgerFileType,
			testUtils.BadgerIMType,
			testUtils.DefraIMType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.DeleteCollection{
				ActiveOnly:    true,
				TransactionID: immutable.Some(1),
				Names:         []string{"Users"},
			},
			// Without committing, the collection data should still exist. Verify via
			// GetCollections which reads from a separate transaction.
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteCollection_DeleteBothCollections_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
					type Books {
						title: String
					}
				`,
			},
			&action.DeleteCollection{
				ActiveOnly: true,
				Names:      []string{"Users"},
			},
			&action.DeleteCollection{
				ActiveOnly: true,
				Names:      []string{"Books"},
			},
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteCollection_ReferencedByRelation_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						books: [Books]
					}
					type Books {
						title: String
						author: Users
					}
				`,
			},
			&action.DeleteCollection{
				ActiveOnly:    true,
				Names:         []string{"Users"},
				ExpectedError: "cannot remove a collection while another field references it",
			},
			// The failed delete must roll back atomically: both collections still exist.
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Books",
						IsMaterialized: true,
						IsActive:       true,
					},
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestDeleteCollection_ReferencedByRelation_OtherSide_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						books: [Books]
					}
					type Books {
						title: String
						author: Users
					}
				`,
			},
			&action.DeleteCollection{
				ActiveOnly:    true,
				Names:         []string{"Books"},
				ExpectedError: "cannot remove a collection while another field references it",
			},

			// The failed delete must roll back atomically: both collections still exist.
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Books",
						IsMaterialized: true,
						IsActive:       true,
					},
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Multiple unrelated collections can be deleted in a single call.
func TestDeleteCollection_MultipleCollections_AtomicallyInOneCall(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
					type Books {
						title: String
					}
				`,
			},
			&action.DeleteCollection{
				ActiveOnly: true,
				Names:      []string{"Users", "Books"},
			},
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// The whole point of multi-delete: two collections referencing each other cannot be
// deleted one at a time (see the relation-error tests above), but passing both to a
// single DeleteCollection call succeeds because the underlying patch is atomic.
func TestDeleteCollection_BothRelatedCollections_InOneCall_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						books: [Books]
					}
					type Books {
						title: String
						author: Users
					}
				`,
			},
			&action.DeleteCollection{
				ActiveOnly: true,
				Names:      []string{"Users", "Books"},
			},
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// If any name in the list is unknown, the whole call fails and nothing is deleted.
func TestDeleteCollection_MixedValidAndInvalidName_RollsBack(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.DeleteCollection{
				ActiveOnly:    true,
				Names:         []string{"Users", "NonExistent"},
				ExpectedError: "collection not found",
			},
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       true,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Default delete (ActiveOnly == false) removes every version of the named collection,
// active head and all earlier (inactive) versions.
func TestDeleteCollection_Default_RemovesAllVersions(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			// Adding a field promotes the new version to active and demotes the original
			// to an inactive earlier version.
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} }
					]
				`,
			},
			// Default behaviour: every version is removed including the inactive earlier one.
			&action.DeleteCollection{
				Names: []string{"Users"},
			},
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// With ActiveOnly the earlier version remains after the active head is removed; it is
// effectively a rollback to the previous version (which is promoted to active).
func TestDeleteCollection_ActiveOnly_LeavesEarlierVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} }
					]
				`,
			},
			&action.DeleteCollection{
				ActiveOnly: true,
				Names:      []string{"Users"},
			},
			// One version of Users still exists after the active head was removed; it
			// was the earlier (inactive) version and remains inactive.
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       false,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Default delete across multiple collections each with multiple versions removes all of them.
func TestDeleteCollection_Default_MultipleCollectionsWithMultipleVersions_RemovesAllVersions(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
					type Books {}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} }
					]
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Books/Fields/-", "value": {"Name": "title", "Kind": "String"} }
					]
				`,
			},
			&action.DeleteCollection{
				Names: []string{"Users", "Books"},
			},
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
