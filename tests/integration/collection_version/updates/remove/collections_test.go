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
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestColVersionUpdateRemoveCollections_ByID(t *testing.T) {
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
						{
							"op": "remove",
							"path": "/bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
						}
					]
				`,
			},
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{},
			},
			&action.Request{
				Request: `mutation {
						add_Users(input:{}) {
							name
						}
					}`,
				ExpectedError: `Cannot query field "add_Users" on type "Mutation".`,
			},
			&action.Request{
				Request: `mutation {
						update_Users(input:{}) {
							name
						}
					}`,
				ExpectedError: `Cannot query field "update_Users" on type "Mutation".`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				ExpectedError: `Cannot query field "Users" on type "Query".`,
			},
			&action.SubscriptionRequest{
				Request: `subscription {
					Users {
						name
					}
				}`,
				ExpectedError: `Cannot query field "Users" on type "Subscription".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateRemoveCollections_ByName(t *testing.T) {
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
						{
							"op": "remove",
							"path": "/Users"
						}
					]
				`,
			},
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{},
			},
			&action.Request{
				Request: `mutation {
						add_Users(input:{}) {
							name
						}
					}`,
				ExpectedError: `Cannot query field "add_Users" on type "Mutation".`,
			},
			&action.Request{
				Request: `mutation {
						update_Users(input:{}) {
							name
						}
					}`,
				ExpectedError: `Cannot query field "update_Users" on type "Mutation".`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				ExpectedError: `Cannot query field "Users" on type "Query".`,
			},
			&action.SubscriptionRequest{
				Request: `subscription {
					Users {
						name
					}
				}`,
				ExpectedError: `Cannot query field "Users" on type "Subscription".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateRemoveCollectionWithData(t *testing.T) {
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
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/Users"
						}
					]
				`,
				ExpectedError: "cannot delete a collection that has documents, first delete the documents and then delete the version",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateRemoveCollectionWithSoftDeletedData(t *testing.T) {
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
			// Soft delete the document, preserving it in the datastore.
			testUtils.DeleteDoc{
				CollectionID: 0,
				DocID:        0,
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
				ExpectedError: "cannot delete a collection that has documents, first delete the documents and then delete the version",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateCopyCollectionAddFieldRemoveOriginalCollection(t *testing.T) {
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
				// The net result of the `copy` followed by the `remove` is zero due to the way the internals are currently
				// coded.
				Patch: `
					[
						{
							"op": "copy",
							"from": "/Users",
							"path": "/UsersV2"
						},
						{
							"op": "add", "path": "/UsersV2/Fields/-", "value": {"Name": "email", "Kind": 11}
						},
						{
							"op": "remove",
							"path": "/Users"
						}
					]
				`,
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
					},
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
						}),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateAddFieldRemoveOriginalCollection_SamePatch(t *testing.T) {
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
				// Because the `remove` operation is applied before the new versionID is set by `add`, the end result
				// of this patch is the deletion of the collection.
				Patch: `
					[
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11}
						},
						{
							"op": "remove",
							"path": "/bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
						}
					]
				`,
			},
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateAddFieldRemoveOriginalCollection_DifferentPatches(t *testing.T) {
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
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11}
						}
					]
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
						}
					]
				`,
				ExpectedError: "cannot delete a version that is used by a newer version, first delete the new version",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateAddFieldRemoveNewCollection_DifferentPatches(t *testing.T) {
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
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11}
						}
					]
				`,
			},
			&action.PatchCollection{
				// Remove the active version, leaving the collection un-queriable
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
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       false,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateAddFieldRemoveNewCollectionAndActivateOriginal(t *testing.T) {
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
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11}
						}
					]
				`,
			},
			&action.PatchCollection{
				// Remove the active version, and activate the original verison
				Patch: `
					[
						{
							"op": "remove",
							"path": "/Users"
						},
						{
							"op": "replace",
							"path": "/bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu/IsActive",
							"value": true
						}
					]
				`,
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
			&action.AddDoc{
				// It is important that this test adds and queries a document as it is possible
				// for the code to be written in a way that erroneously deletes the field short ids
				// for fields that existed for non-deleted versions.
				DocMap: map[string]any{
					"name": "John",
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

func TestColVersionUpdateAddFieldRemoveMultipleNewCollection_FirstAndLast(t *testing.T) {
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
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11}
						}
					]
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "postCode", "Kind": 11}
						}
					]
				`,
			},
			&action.PatchCollection{
				// Remove the first and last versions
				Patch: `
					[
						{
							"op": "remove",
							"path": "/bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
						},
						{
							"op": "remove",
							"path": "/Users"
						}
					]
				`,
				ExpectedError: "cannot delete a version that is used by a newer version, first delete the new version",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateAddFieldRemoveMultipleNewCollection_FirstAndMiddle(t *testing.T) {
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
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11}
						}
					]
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "postCode", "Kind": 11}
						}
					]
				`,
			},
			&action.PatchCollection{
				// Remove the first and middle versions
				Patch: `
					[
						{
							"op": "remove",
							"path": "/bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
						},
						{
							"op": "remove",
							"path": "/bafyreigvzkfdc4y2ppvvpmmdw3t7kv4nd5dgfh5jfytef3kbzem6po55zu"
						}
					]
				`,
				ExpectedError: "cannot delete a version that is used by a newer version, first delete the new version",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateAddFieldRemoveMultipleNewCollection_MiddleAndLast(t *testing.T) {
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
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11}
						}
					]
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "postCode", "Kind": 11}
						}
					]
				`,
			},
			&action.PatchCollection{
				// Remove the middle and last versions
				Patch: `
					[
						{
							"op": "remove",
							"path": "/bafyreigvzkfdc4y2ppvvpmmdw3t7kv4nd5dgfh5jfytef3kbzem6po55zu"
						},
						{
							"op": "remove",
							"path": "/Users"
						}
					]
				`,
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsMaterialized: true,
						IsActive:       false,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Removing a single collection via patch fails when another collection holds a relation
// reference to it. The schema rebuild cannot resolve the dangling reference and aborts
// the transaction, leaving both collections intact.
func TestColVersionUpdateRemoveCollection_ReferencedByRelation_ReturnsError(t *testing.T) {
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
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "remove", "path": "/Users" }
					]
				`,
				ExpectedError: "cannot remove a collection while another field references it",
			},
			// Transaction rolled back: both collections still exist.
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

// Same as above but removing the other side of the bidirectional relation.
func TestColVersionUpdateRemoveCollection_ReferencedByRelation_OtherSide_ReturnsError(t *testing.T) {
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
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "remove", "path": "/Books" }
					]
				`,
				ExpectedError: "cannot remove a collection while another field references it",
			},
			// Transaction rolled back: both collections still exist.
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

// A single patch with both remove ops succeeds because the net result has no dangling
// references. This is the escape hatch for deleting circularly-related collections.
func TestColVersionUpdateRemoveBothRelatedCollections_Succeeds(t *testing.T) {
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
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "remove", "path": "/Users" },
						{ "op": "remove", "path": "/Books" }
					]
				`,
			},
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateRemoveCollections_ConcurrentWrite(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// The other client types return different errors when occasionally executing the `CreateDoc`
			// action.
			state.GoClientType,
		}),
		SupportedDatabaseTypes: immutable.Some([]state.DatabaseType{
			// LevelDB is not supported for this test as the test opens multiple transactions at
			// the same time.
			testUtils.BadgerIMType,
			testUtils.BadgerFileType,
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
			&action.Async{
				// todo - we also need to test this with explicit transactions both async and sync
				// https://github.com/sourcenetwork/defradb/issues/4476
				Child: &action.PatchCollection{
					// If the create call completes before the patch starts this will error - skip the test
					// when this happens as it is unrecoverable and rare.  The production code in such a
					// scenario is behaving correctly.
					SkipTestOnError: description.ErrCannotDeleteCollectionWithDocs,
					Patch: `
						[
							{
								"op": "remove",
								"path": "/Users"
							}
						]
					`,
				},
			},
			&action.AddDoc{
				DoNotWaitForEvent: true,
				DocMap: map[string]any{
					"name": "John",
				},
				// This error can occur if the create-doc call starts after the patch collection call (mostly)
				// completes, it is uncommon for this to happen, but it does sometimes, especially on slower
				// machines.  It is correct behaviour, but is not the scenario that this test is asserting.
				IgnoreError: "collection not found",
			},
			&action.Await{},
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{},
			},
			&action.Request{
				Request: `query {
					_commits {
						cid
					}
				}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
