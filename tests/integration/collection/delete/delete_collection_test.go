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

// Deleting multiple collections inside a transaction that is never committed must
// leave all of them intact.
func TestDeleteCollection_MultipleCollections_WithTransactionWithoutCommit_StillExist(t *testing.T) {
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
					type Books {
						title: String
					}
				`,
			},
			&action.DeleteCollection{
				TransactionID: immutable.Some(1),
				Names:         []string{"Users", "Books"},
			},

			// Without committing, the collection data should still exist.
			// Verify via GetCollections which reads from a separate transaction.
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
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
func TestDeleteCollection_MultipleCollections_AtomicallyInOneCall_Succeeds(t *testing.T) {
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
func TestDeleteCollection_BothRelatedCollections_ActiveOnly_InOneCall_Succeeds(t *testing.T) {
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
				FilterOptions:   options.GetCollections().SetGetInactive(true),
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
				ActiveOnly: false,
				Names:      []string{"Users", "Books"},
			},
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetGetInactive(true),
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

func TestDeleteCollection_BackToBackPatchAndDelete_AllVersionsRemoved(t *testing.T) {
	// usersV1 is the initial `type Users { name: String }`.
	const usersV1 = "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
	// usersV2 is the patched schema with the `age: Int` field added.
	const usersV2 = "bafyreics7adsddesun4kqqotr6g6c6ld2t7djlwcbrm4ftbhru3ayindy4"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			// First patch: creates v2 active; v1 inactive.
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "age", "Kind": "Int"} }
					]
				`,
			},
			// Second patch: creates v3 active; v2 and v1 inactive.
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
					]
				`,
			},
			// Delete the active head v3, leaving v2 and v1 inactive in storage.
			&action.DeleteCollection{
				ActiveOnly: true,
				Names:      []string{"Users"},
			},
			// Promote v2 to active so the next active-only delete targets it.
			testUtils.SetActiveCollectionVersion{
				VersionID: usersV2,
			},
			// Delete the active head v2, leaving v1 inactive in storage.
			&action.DeleteCollection{
				ActiveOnly: true,
				Names:      []string{"Users"},
			},
			// Promote v1 to active so the final default delete can resolve it by name.
			testUtils.SetActiveCollectionVersion{
				VersionID: usersV1,
			},
			// Final default delete removes the lone v1.
			&action.DeleteCollection{
				Names: []string{"Users"},
			},
			// After the back-to-back sequence, no version of Users should remain.
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

// After both related collections are removed atomically, queries against either
// must fail at the GQL layer (the types are not in the active schema anymore).
func TestDeleteCollection_RelatedPair_QueriesAgainstSurvivor(t *testing.T) {
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
			// Atomic delete of both collections.
			&action.DeleteCollection{
				Names: []string{"Users", "Books"},
			},
			// Querying either collection now must error with "Cannot query field".
			&action.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				ExpectedError: `Cannot query field "Users" on type "Query".`,
			},
			&action.Request{
				Request: `query {
					Books {
						title
					}
				}`,
				ExpectedError: `Cannot query field "Books" on type "Query".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// A collection that self-references via a single object field is deletable
// using the default (all-versions) path.
func TestDeleteCollection_SelfReferencingOneToOne_Default_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Person {
						name: String
						parent: Person
					}
				`,
			},
			&action.DeleteCollection{
				Names: []string{"Person"},
			},
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{},
			},
			&action.Request{
				Request: `query {
					Person {
						name
					}
				}`,
				ExpectedError: `Cannot query field "Person" on type "Query".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Self-reference via an array (one-to-many to self). DefraDB requires both sides of a one-to-many
// self-relation to be declared explicitly in the SDL: a `parent: Person @primary` field paired
// with the `children: [Person]` array side. If we didn't do that and only had the array side alone
// then it would have been rejected at `AddCollection` step.
func TestDeleteCollection_SelfReferencingOneToMany_Default_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Person {
						name: String
						parent: Person @primary
						children: [Person]
					}
				`,
			},
			&action.DeleteCollection{
				Names: []string{"Person"},
			},
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{},
			},
			&action.Request{
				Request: `query {
					Person {
						name
					}
				}`,
				ExpectedError: `Cannot query field "Person" on type "Query".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Self-reference + --active-only on a multi-version collection.
func TestDeleteCollection_SelfReferencing_ActiveOnly_LeavesEarlierVersion(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Person {
						parent: Person
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Person/Fields/-", "value": {"Name": "name", "Kind": "String"} }
					]
				`,
			},
			&action.DeleteCollection{
				ActiveOnly: true,
				Names:      []string{"Person"},
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Person",
						IsMaterialized: true,
						IsActive:       false,
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// After an --active-only delete of the only-active version, the inactive earlier version
// remains but is not query-accessible because the active schema is empty.
// The user can recover by promoting the earlier version with [SetActiveCollectionVersion].
func TestDeleteCollection_ActiveOnly_FollowedBySetActive_RestoresQueries(t *testing.T) {
	// CID for the v1 Users schema (single `name: String` field).
	const usersV1 = "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "age", "Kind": "Int"} }
					]
				`,
			},
			// --active-only delete removes Users-v2 (the active head); Users-v1 remains
			// in storage but inactive, so it isn't reachable via queries.
			&action.DeleteCollection{
				ActiveOnly: true,
				Names:      []string{"Users"},
			},
			// The delete only removed the active head; v1 must remain, inactive.
			// Asserted here so a dropped-v1 regression is caught at the delete step.
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						VersionID:      usersV1,
						IsMaterialized: true,
						IsActive:       false,
					},
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
			// Promote the earlier (v1) version back to active. After this Users is in the
			// active schema again, exposing only v1's fields.
			testUtils.SetActiveCollectionVersion{
				VersionID: usersV1,
			},
			&action.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
			// v1 doesn't have the `age` field that was added in v2 - querying it must error.
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

// Default delete (all versions) of one collection while an unrelated collection has multiple
// inactive versions. Ensures the delete only touches the named collection's CollectionID.
func TestDeleteCollection_UnrelatedCollectionWithMultipleVersions_PreservedAfterDelete(t *testing.T) {
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
			// Create multiple versions on Books.
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Books/Fields/-", "value": {"Name": "year", "Kind": "Int"} }
					]
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Books/Fields/-", "value": {"Name": "isbn", "Kind": "String"} }
					]
				`,
			},
			// Delete Users. Books versions must all survive untouched.
			&action.DeleteCollection{
				Names: []string{"Users"},
			},
			// Users must be entirely gone.
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetCollectionName("Users").SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{},
			},
			// Books queries against the latest schema must still work, proving the active
			// Books version (and the field added in the second patch) survived the delete.
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
			// The active query above only proves v3 (the active head) survived; this
			// proves the two inactive earlier versions survived the delete too.
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetCollectionName("Books").SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Books",
						IsMaterialized: true,
						IsActive:       false,
					},
					{
						Name:           "Books",
						IsMaterialized: true,
						IsActive:       false,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "bafyreibb6jablg4srs2juy5u7uuitiee23toxgqei33i63yh6qfhdju3c4",
						}),
					},
					{
						Name:           "Books",
						IsMaterialized: true,
						IsActive:       true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "bafyreie67mjmn6k2zvfhsoudrnlagfoe256t7egpxubkehpb3dkdqn6s5y",
						}),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Collection with an index attached gets deleted properly.
func TestDeleteCollection_WithSecondaryIndex_RemovesIndex(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.NewIndex{
				CollectionID: 0,
				FieldName:    "name",
			},
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

// A view that depends on a collection does NOT block deletion of that source
// collection. This is by design: views are derived data, not structural
// references. They materialize on demand from their source query, so it is safe
// to let them outlive the source. The view's query now resolves against a missing
// collection and surfaces a "collection not found" error (view has gone stale).
func TestDeleteCollection_WithViewDepending_ViewBecomesStale(t *testing.T) {
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
				Query: `Users { name }`,
				SDL: `
					type UsersByName {
						name: String
					}
				`,
			},
			// Sanity check: while the source exists, the view is queryable and
			// returns an empty result set (no docs in Users yet).
			&action.Request{
				Request: `query {
					UsersByName {
						name
					}
				}`,
				Results: map[string]any{
					"UsersByName": []map[string]any{},
				},
			},
			// The delete succeeds: the view does not hold the source hostage.
			// (If the same dependency had been a relation field on another
			// collection, this step would have errored)
			&action.DeleteCollection{
				Names: []string{"Users"},
			},
			&action.Request{
				Request: `query {
					UsersByName {
						name
					}
				}`,
				ExpectedError: "collection not found",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// An empty string in the names slice should produce a clear error, not a silent
// skip or a different downstream failure.
func TestDeleteCollection_NameContainsEmptyString_ReturnsError(t *testing.T) {
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
				Names:         []string{"Users", ""},
				ExpectedError: "collection name can't be empty",
			},
			// Users must still exist.
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

// Repeating the same name in the names slice should be deduplicated and
// behave the same as a single-name request.
func TestDeleteCollection_DuplicateNames_DedupesAndSucceeds(t *testing.T) {
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
				Names: []string{"Users", "Users", "Users"},
			},
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
