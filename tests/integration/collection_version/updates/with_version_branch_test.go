// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package updates

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionUpdates_WithBranchingCollection(t *testing.T) {
	collectionVersion1ID := "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
	collectionVersion2ID := "bafyreigvzkfdc4y2ppvvpmmdw3t7kv4nd5dgfh5jfytef3kbzem6po55zu"
	collectionVersion3ID := "bafyreifwalt5gom7ldime4phszmbxymn5jrtkx33ujw7ovvjmdzpat5yzm"

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
				// The second collection version will not be set as the active version, leaving the initial version active
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			&action.PatchCollection{
				// The third collection version will be set as the active version, going from version 1 to 3
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
					]
				`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				// The email field is not queriable
				ExpectedError: `Cannot query field "email" on type "Users".`,
			},
			&action.GetCollections{
				// The second collection version is present in the system, with the email field
				FilterOptions: options.GetCollections().SetVersionID(collectionVersion2ID),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						VersionID:      collectionVersion2ID,
						CollectionID:   collectionVersion1ID,
						IsActive:       false,
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: collectionVersion1ID,
						}),
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "email",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
			&action.Request{
				// The phone field is queriable
				Request: `query {
					Users {
						name
						phone
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
			&action.GetCollections{
				// The third collection version is present in the system, with the phone field
				FilterOptions: options.GetCollections().SetVersionID(collectionVersion3ID),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						VersionID:      collectionVersion3ID,
						CollectionID:   collectionVersion1ID,
						IsActive:       true,
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: collectionVersion1ID,
						}),
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "phone",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						// The original collection version is present, it has no source and is inactive (has no name).
						VersionID:      collectionVersion1ID,
						IsMaterialized: true,
						Name:           "Users",
					},
					{
						// The collection version for collection version 3 is present and is active, it also has the first collection
						// as source.
						Name:           "Users",
						VersionID:      collectionVersion3ID,
						IsMaterialized: true,
						IsActive:       true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: collectionVersion1ID,
						}),
					},
					{
						// The collection version for collection version 2 is present, it has the first collection as a source
						// and is inactive.
						Name:           "Users",
						VersionID:      collectionVersion2ID,
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: collectionVersion1ID,
						}),
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdates_WithPatchOnBranchedCollection(t *testing.T) {
	collectionVersion1ID := "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
	collectionVersion2ID := "bafyreigvzkfdc4y2ppvvpmmdw3t7kv4nd5dgfh5jfytef3kbzem6po55zu"
	collectionVersion3ID := "bafyreifwalt5gom7ldime4phszmbxymn5jrtkx33ujw7ovvjmdzpat5yzm"
	collectionVersion4ID := "bafyreibuscrpd27xb2zelovaid6souccvac5rkl4xrvjowe3jpfhormr6e"

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
				// The second collection version will not be set as the active version, leaving the initial version active
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			&action.PatchCollection{
				// The third collection version will be set as the active version, going from version 1 to 3
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
					]
				`,
			},
			&action.PatchCollection{
				// The fourth collection version will be set as the active version, going from version 3 to 4
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "discordName", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
					]
				`,
			},
			&action.Request{
				// The phone and discordName fields are queriable
				Request: `query {
					Users {
						name
						phone
						discordName
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
			&action.GetCollections{
				// The fourth collection version is present in the system, with the phone and discordName field
				FilterOptions: options.GetCollections().SetVersionID(collectionVersion4ID),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						VersionID:      collectionVersion4ID,
						CollectionID:   collectionVersion1ID,
						IsActive:       true,
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: collectionVersion3ID,
						}),
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "phone",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "discordName",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						// The collection version for collection version 4 is present and is active, it also has the third collection
						// as source.
						Name:           "Users",
						VersionID:      collectionVersion4ID,
						IsMaterialized: true,
						IsActive:       true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: collectionVersion3ID,
						}),
					},
					{
						// The original collection version is present, it has no source and is inactive
						Name:           "Users",
						VersionID:      collectionVersion1ID,
						IsMaterialized: true,
						IsActive:       false,
					},
					{
						// The collection version for collection version 3 is present and inactive, it has the first collection
						// as source.
						Name:           "Users",
						VersionID:      collectionVersion3ID,
						IsMaterialized: true,
						IsActive:       false,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: collectionVersion1ID,
						}),
					},
					{
						// The collection version for collection version 2 is present, it has the first collection as a source
						// and is inactive.
						Name:           "Users",
						VersionID:      collectionVersion2ID,
						IsMaterialized: true,
						IsActive:       false,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: collectionVersion1ID,
						}),
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdates_WithBranchingCollectionAndSetActiveCollectionToOtherBranch(t *testing.T) {
	collectionVersion1ID := "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
	collectionVersion2ID := "bafyreigvzkfdc4y2ppvvpmmdw3t7kv4nd5dgfh5jfytef3kbzem6po55zu"
	collectionVersion3ID := "bafyreifwalt5gom7ldime4phszmbxymn5jrtkx33ujw7ovvjmdzpat5yzm"

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
				// The second collection version will not be set as the active version, leaving the initial version active
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			&action.PatchCollection{
				// The third collection version will be set as the active version, going from version 1 to 3
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
				// Set the second collection version to be active
				VersionID: collectionVersion2ID,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				// The email field is queriable
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
			&action.Request{
				Request: `query {
					Users {
						name
						phone
					}
				}`,
				// The phone field is not queriable
				ExpectedError: `Cannot query field "phone" on type "Users".`,
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						// The original collection version is present, it has no source and is inactive.
						Name:           "Users",
						VersionID:      collectionVersion1ID,
						IsMaterialized: true,
						IsActive:       false,
					},
					{
						// The collection version for collection version 3 is present and is inactive, it also has the first collection
						// as source.
						Name:           "Users",
						VersionID:      collectionVersion3ID,
						IsMaterialized: true,
						IsActive:       false,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: collectionVersion1ID,
						}),
					},
					{
						// The collection version for collection version 2 is present and is active, it has the first collection as a source
						Name:           "Users",
						VersionID:      collectionVersion2ID,
						IsMaterialized: true,
						IsActive:       true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: collectionVersion1ID,
						}),
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdates_WithBranchingCollectionAndSetActiveCollectionToOtherBranchThenPatch(t *testing.T) {
	collectionVersion1ID := "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"
	collectionVersion2ID := "bafyreigvzkfdc4y2ppvvpmmdw3t7kv4nd5dgfh5jfytef3kbzem6po55zu"
	collectionVersion3ID := "bafyreifwalt5gom7ldime4phszmbxymn5jrtkx33ujw7ovvjmdzpat5yzm"
	collectionVersion4ID := "bafyreibuscrpd27xb2zelovaid6souccvac5rkl4xrvjowe3jpfhormr6e"

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
				// The second collection version will not be set as the active version, leaving the initial version active
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			&action.PatchCollection{
				// The third collection version will be set as the active version, going from version 1 to 3
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
				// Set the second collection version to be active
				VersionID: collectionVersion2ID,
			},
			&action.PatchCollection{
				// The fourth collection version will be set as the active version, going from version 2 to 4
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "discordName", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
					]
				`,
			},
			&action.Request{
				// The email and discordName fields are queriable
				Request: `query {
					Users {
						name
						email
						discordName
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
			&action.GetCollections{
				// The fourth collection version is present in the system, with the email and discordName field
				FilterOptions: options.GetCollections().SetVersionID(collectionVersion4ID),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						VersionID:      collectionVersion4ID,
						CollectionID:   collectionVersion1ID,
						IsActive:       true,
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: collectionVersion2ID,
						}),
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.NONE_CRDT,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "email",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "discordName",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						// The collection version for collection version 4 is present and is active, it also has the second collection
						// as source.
						Name:           "Users",
						VersionID:      collectionVersion4ID,
						IsMaterialized: true,
						IsActive:       true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: collectionVersion2ID,
						}),
					},
					{
						// The original collection version is present, it has no source and is inactive.
						Name:           "Users",
						VersionID:      collectionVersion1ID,
						IsMaterialized: true,
						IsActive:       false,
					},
					{
						// The collection version for collection version 3 is present and inactive, it has the first collection
						// as source.
						Name:           "Users",
						VersionID:      collectionVersion3ID,
						IsMaterialized: true,
						IsActive:       false,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: collectionVersion1ID,
						}),
					},
					{
						// The collection version for collection version 2 is present, it has the first collection as a source
						// and is inactive.
						Name:           "Users",
						VersionID:      collectionVersion2ID,
						IsMaterialized: true,
						IsActive:       false,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: collectionVersion1ID,
						}),
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdates_WithBranchingCollectionAndGetCollectionAtVersion(t *testing.T) {
	collectionVersion1ID := "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu"

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
				// The second collection version will not be set as the active version, leaving the initial version active
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
					]
				`,
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetVersionID(collectionVersion1ID),
				ExpectedResults: []client.CollectionVersion{
					{
						// The original collection version is present, it has no source and is inactive.
						Name:           "Users",
						VersionID:      collectionVersion1ID,
						IsMaterialized: true,
						IsActive:       false,
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
