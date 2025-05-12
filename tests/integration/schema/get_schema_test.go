// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestGetSchema_GivenNonExistantSchemaVersionID_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					VersionID: immutable.Some("does not exist"),
				},
				ExpectedError: "key not found",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetSchema_GivenNoSchemaReturnsEmptySet(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetSchema_GivenNoSchemaGivenUnknownRoot(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					CollectionID: immutable.Some("does not exist"),
				},
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetSchema_GivenNoSchemaGivenUnknownName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					Name: immutable.Some("does not exist"),
				},
				ExpectedResults: []client.CollectionVersion{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetSchema_ReturnsAllSchema(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
			},
			&action.AddSchema{
				Schema: `
					type Books {}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
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
						IsActive:       false,
						IsMaterialized: true,
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

func TestGetSchema_ReturnsSchemaForGivenRoot(t *testing.T) {
	usersSchemaVersion1ID := "bafyreihdbjfazsx5vq2tpzedqdktrjyn6lq22qle7el2s42b3q4zpxmwqq"
	usersSchemaVersion2ID := "bafyreidfzu2x6i4akqlmt5lloaeflwep4ykq2f4unwm2utkmrkgpfyf7bi"

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
			},
			&action.AddSchema{
				Schema: `
					type Books {}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
					CollectionID:    immutable.Some(usersSchemaVersion1ID),
				},
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						CollectionID:   usersSchemaVersion1ID,
						VersionID:      usersSchemaVersion1ID,
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
						CollectionID:   usersSchemaVersion1ID,
						VersionID:      usersSchemaVersion2ID,
						IsActive:       false,
						IsMaterialized: true,
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

func TestGetSchema_ReturnsSchemaForGivenName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
			},
			&action.AddSchema{
				Schema: `
					type Books {}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					Name:            immutable.Some("Users"),
					IncludeInactive: immutable.Some(true),
				},
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						IsActive:       false,
						IsMaterialized: true,
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
