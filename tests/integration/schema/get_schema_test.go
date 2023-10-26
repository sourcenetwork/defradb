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

func TestGetSchema_ReturnsAllSchema(t *testing.T) {
	usersSchemaVersion1ID := "bafkreickgf3nbjaairxkkqawmrv7fafaafyccl4qygqeveagisdn42eohu"
	usersSchemaVersion2ID := "bafkreicseqwxooxo2wf2bgzdalwtm2rtsj7x4mgsir4rp4htmpnwnffwre"
	booksSchemaVersion1ID := "bafkreigbfibfn7g6neen2gghc54dzocexefi7vshc3opgvy6j7jflar2nm"

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
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
							},
						},
					},
					{
						Name:      "Users",
						Root:      usersSchemaVersion1ID,
						VersionID: usersSchemaVersion2ID,
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								ID:   1,
								Kind: client.FieldKind_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
					{
						Name:      "Books",
						Root:      booksSchemaVersion1ID,
						VersionID: booksSchemaVersion1ID,
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
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
	usersSchemaVersion1ID := "bafkreickgf3nbjaairxkkqawmrv7fafaafyccl4qygqeveagisdn42eohu"
	usersSchemaVersion2ID := "bafkreicseqwxooxo2wf2bgzdalwtm2rtsj7x4mgsir4rp4htmpnwnffwre"

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
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
							},
						},
					},
					{
						Name:      "Users",
						Root:      usersSchemaVersion1ID,
						VersionID: usersSchemaVersion2ID,
						Fields: []client.FieldDescription{
							{
								Name: "_key",
								Kind: client.FieldKind_DocKey,
								Typ:  client.LWW_REGISTER,
							},
							{
								Name: "name",
								ID:   1,
								Kind: client.FieldKind_STRING,
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
