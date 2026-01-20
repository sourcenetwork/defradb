// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package remove

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateRemoveCollections_ByID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
						}
					]
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{},
			},
			testUtils.Request{
				Request: `mutation {
						create_Users(input:{}) {
							name
						}
					}`,
				ExpectedError: `Cannot query field "create_Users" on type "Mutation".`,
			},
			testUtils.Request{
				Request: `mutation {
						update_Users(input:{}) {
							name
						}
					}`,
				ExpectedError: `Cannot query field "update_Users" on type "Mutation".`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				ExpectedError: `Cannot query field "Users" on type "Query".`,
			},
			testUtils.SubscriptionRequest{
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
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/Users"
						}
					]
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{},
			},
			testUtils.Request{
				Request: `mutation {
						create_Users(input:{}) {
							name
						}
					}`,
				ExpectedError: `Cannot query field "create_Users" on type "Mutation".`,
			},
			testUtils.Request{
				Request: `mutation {
						update_Users(input:{}) {
							name
						}
					}`,
				ExpectedError: `Cannot query field "update_Users" on type "Mutation".`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				ExpectedError: `Cannot query field "Users" on type "Query".`,
			},
			testUtils.SubscriptionRequest{
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
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.PatchCollection{
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
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"name": "John",
				},
			},
			// Soft delete the document, preserving it in the datastore.
			testUtils.DeleteDoc{
				CollectionID: 0,
				DocID:        0,
			},
			testUtils.PatchCollection{
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
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
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
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
				},
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
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
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
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
				},
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateAddFieldRemoveOriginalCollection_DifferentPatches(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11}
						}
					]
				`,
			},
			testUtils.PatchCollection{
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
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11}
						}
					]
				`,
			},
			testUtils.PatchCollection{
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
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
				},
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
			testUtils.Request{
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
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11}
						}
					]
				`,
			},
			testUtils.PatchCollection{
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
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
				},
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
			testUtils.CreateDoc{
				// It is important that this test creates and queries a document as it is possible
				// for the code to be written in a way that erroneously deletes the field short ids
				// for fields that existed for non-deleted versions.
				DocMap: map[string]any{
					"name": "John",
				},
			},
			testUtils.Request{
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
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11}
						}
					]
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "postCode", "Kind": 11}
						}
					]
				`,
			},
			testUtils.PatchCollection{
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
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11}
						}
					]
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "postCode", "Kind": 11}
						}
					]
				`,
			},
			testUtils.PatchCollection{
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
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11}
						}
					]
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "add", "path": "/Users/Fields/-", "value": {"Name": "postCode", "Kind": 11}
						}
					]
				`,
			},
			testUtils.PatchCollection{
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
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
				},
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
