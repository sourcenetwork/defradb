// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestGetCollectionVersion_GivenNonExistantCollectionVersionID_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetVersionID("does not exist"),
				ExpectedError: "collection not found",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetCollectionVersion_GivenNoCollectionReturnsEmptySet(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.GetCollections{
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetCollectionVersion_GivenNoCollectionGivenUnknownRoot(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetCollectionID("does not exist"),
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetCollectionVersion_GivenNoCollectionGivenUnknownName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.GetCollections{
				FilterOptions:   options.GetCollections().SetCollectionName("does not exist"),
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetCollectionVersion_ReturnsAllCollections(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.AddCollection{
				SDL: `
					type Books {}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Books",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name:           "Users",
						IsActive:       false,
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna",
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
						},
					},
					{
						Name:           "Users",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetCollectionVersion_ReturnsCollectionForGivenRoot(t *testing.T) {
	usersCollectionVersion1ID := "bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna"
	usersCollectionVersion2ID := "bafyreieqhzanpek5ssb7ofi3qelbvl2nwh6s7x3w2mlzbcnqaqol3elltq"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.AddCollection{
				SDL: `
					type Books {}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetGetInactive(true).SetCollectionID(usersCollectionVersion1ID),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						CollectionID:   usersCollectionVersion1ID,
						VersionID:      usersCollectionVersion1ID,
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name:           "Users",
						CollectionID:   usersCollectionVersion1ID,
						VersionID:      usersCollectionVersion2ID,
						IsActive:       false,
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: usersCollectionVersion1ID,
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
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetCollectionVersion_ReturnsCollectionForGivenName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.AddCollection{
				SDL: `
					type Books {}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			&action.GetCollections{
				FilterOptions: options.GetCollections().SetCollectionName("Users").SetGetInactive(true),
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsActive:       false,
						IsMaterialized: true,
						PreviousVersion: immutable.Some(client.CollectionSource{
							SourceCollectionID: "bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna",
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
						},
					},
					{
						Name:           "Users",
						IsActive:       true,
						IsMaterialized: true,
						Fields: []client.CollectionFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
