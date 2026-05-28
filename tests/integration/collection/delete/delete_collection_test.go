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

// Adding a collection with the same name after a default delete creates a
// fresh collection (a new CollectionID), not a resurrection of the previous one.
func TestDeleteCollection_ThenAddSameName_IsFreshCollection(t *testing.T) {
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
				Names: []string{"Users"},
			},
			// Re-add with a different shape.
			&action.AddCollection{
				SDL: `
					type Users {
						email: String
					}
				`,
			},
			// Querying email must work - proves the collection was cleanly reset.
			&action.Request{
				Request: `query {
					Users {
						email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
			// And the old field should not exist.
			&action.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				ExpectedError: `Cannot query field "name" on type "Users".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Multi-name delete must clean the collection-definition and field-definition
// heads for every name in the list, not just one. After deleting Users + Books
// in a single call, both names must be reusable for fresh collections - if
// either name's heads were left stale, the corresponding `AddCollection` would
// error.
func TestDeleteCollection_MultiName_ThenAddSameNames_AreFreshCollections(t *testing.T) {
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
				Names: []string{"Users", "Books"},
			},
			// Re-add both names with different shapes.
			&action.AddCollection{
				SDL: `
					type Users {
						email: String
					}
					type Books {
						isbn: String
					}
				`,
			},
			&action.Request{
				Request: `query {
					Users {
						email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
			&action.Request{
				Request: `query {
					Books {
						isbn
					}
				}`,
				Results: map[string]any{
					"Books": []map[string]any{},
				},
			},
			// Old fields must not exist on either re-added collection.
			&action.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				ExpectedError: `Cannot query field "name" on type "Users".`,
			},
			&action.Request{
				Request: `query {
					Books {
						title
					}
				}`,
				ExpectedError: `Cannot query field "title" on type "Books".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// ActiveOnly delete on a single-version collection removes the only version,
// so the headstore-cleanup gate fires (no surviving versions for this name).
// The same-name re-add must therefore work, exactly as in the default-delete
func TestDeleteCollection_ActiveOnly_SingleVersion_ThenAddSameName_IsFreshCollection(t *testing.T) {
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
			// Re-add with a different shape.
			&action.AddCollection{
				SDL: `
					type Users {
						email: String
					}
				`,
			},
			&action.Request{
				Request: `query {
					Users {
						email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
			&action.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				ExpectedError: `Cannot query field "name" on type "Users".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A multi-version collection (Add + PatchCollection) gets default-deleted,
// removing every version; the same-name re-add must then work. This exercises
// the gate's "wait until the last version is gone" behaviour - the per-version
// loop in deleteCollectionVersions must NOT clean heads on the first iteration
// (when v1 is gone but v2 still exists), and MUST clean them on the last.
func TestDeleteCollection_PatchedThenDefaultDeleted_ThenAddSameName_IsFreshCollection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			// Creates v2 active, v1 inactive.
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "age", "Kind": "Int"} }
					]
				`,
			},
			// Default delete removes both v1 and v2.
			&action.DeleteCollection{
				Names: []string{"Users"},
			},
			// Re-add with a different shape - none of the prior fields should leak.
			&action.AddCollection{
				SDL: `
					type Users {
						email: String
					}
				`,
			},
			&action.Request{
				Request: `query {
					Users {
						email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
			&action.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				ExpectedError: `Cannot query field "name" on type "Users".`,
			},
			&action.Request{
				Request: `query {
					Users {
						age
					}
				}`,
				ExpectedError: `Cannot query field "age" on type "Users".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Re-adding the same shape after a delete so the new collection's blocks land
// at the exact same CIDs as the deleted collection's blocks. If the headstore
// had any stale entry pointing at those CIDs, the `AddCollection` path would
// either trip on it or silently accept the stale heads as valid which might
// give the new collection an unrelated history.
func TestDeleteCollection_ThenAddSameShape_IsFreshCollection(t *testing.T) {
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
				Names: []string{"Users"},
			},
			// Re-add with the EXACT same shape.
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			// The new collection must be writable, and a query must return only
			// the doc we just added, no carry-over from the previous instance.
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "Alice",
				},
			},
			&action.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Alice"},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
