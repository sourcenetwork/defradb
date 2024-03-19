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
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestGetSchema_GivenNonExistantSchemaVersionID_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.GetSchema{
				VersionID:     immutable.Some("does not exist"),
				ExpectedError: "datastore: key not found",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetSchema_GivenNoSchemaReturnsEmptySet(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.GetSchema{
				ExpectedResults: []client.SchemaDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetSchema_GivenNoSchemaGivenUnknownRoot(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.GetSchema{
				Root:            immutable.Some("does not exist"),
				ExpectedResults: []client.SchemaDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetSchema_GivenNoSchemaGivenUnknownName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.GetSchema{
				Name:            immutable.Some("does not exist"),
				ExpectedResults: []client.SchemaDescription{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestGetSchema_ReturnsAllSchema(t *testing.T) {
	usersSchemaVersion1ID := "bafkreiaopue5oiqzbszdk265wl6lqkqc44glt2tgjncbwek447slainu7m"
	usersSchemaVersion2ID := "bafkreibuxh4vi3xsob5vx22bn3i5osbkxtimdl2nrs74cqxuf2w3ys2f3y"
	booksSchemaVersion1ID := "bafkreicwmtpmea4gis6lkt46l5evd2xhais36qd5egb2b7mjrqnojbtzja"

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Books {}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} }
					]
				`,
				SetAsDefaultVersion: immutable.Some(false),
			},
			testUtils.GetSchema{
				ExpectedResults: []client.SchemaDescription{
					{
						Name:      "Users",
						Root:      usersSchemaVersion1ID,
						VersionID: usersSchemaVersion1ID,
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name:      "Users",
						Root:      usersSchemaVersion1ID,
						VersionID: usersSchemaVersion2ID,
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
					{
						Name:      "Books",
						Root:      booksSchemaVersion1ID,
						VersionID: booksSchemaVersion1ID,
						Fields: []client.SchemaFieldDescription{
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
	usersSchemaVersion1ID := "bafkreiaopue5oiqzbszdk265wl6lqkqc44glt2tgjncbwek447slainu7m"
	usersSchemaVersion2ID := "bafkreibuxh4vi3xsob5vx22bn3i5osbkxtimdl2nrs74cqxuf2w3ys2f3y"

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Books {}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} }
					]
				`,
				SetAsDefaultVersion: immutable.Some(false),
			},
			testUtils.GetSchema{
				Root: immutable.Some(usersSchemaVersion1ID),
				ExpectedResults: []client.SchemaDescription{
					{
						Name:      "Users",
						Root:      usersSchemaVersion1ID,
						VersionID: usersSchemaVersion1ID,
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name:      "Users",
						Root:      usersSchemaVersion1ID,
						VersionID: usersSchemaVersion2ID,
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.LWW_REGISTER,
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
	usersSchemaVersion1ID := "bafkreiaopue5oiqzbszdk265wl6lqkqc44glt2tgjncbwek447slainu7m"
	usersSchemaVersion2ID := "bafkreibuxh4vi3xsob5vx22bn3i5osbkxtimdl2nrs74cqxuf2w3ys2f3y"

	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
			},
			testUtils.SchemaUpdate{
				Schema: `
					type Books {}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} }
					]
				`,
				SetAsDefaultVersion: immutable.Some(false),
			},
			testUtils.GetSchema{
				Name: immutable.Some("Users"),
				ExpectedResults: []client.SchemaDescription{
					{
						Name:      "Users",
						Root:      usersSchemaVersion1ID,
						VersionID: usersSchemaVersion1ID,
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
						},
					},
					{
						Name:      "Users",
						Root:      usersSchemaVersion1ID,
						VersionID: usersSchemaVersion2ID,
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
								Typ:  client.LWW_REGISTER,
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
