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
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdates_WithBranchingSchema(t *testing.T) {
	schemaVersion1ID := "bafkreiht46o4lakri2py2zw57ed3pdeib6ud6ojlsomgjlrgwh53wl3q4a"
	schemaVersion2ID := "bafkreigdplzukezgpmjs45lw6kwzhtwge4xjzfgm6iodcd32d7kdageply"
	schemaVersion3ID := "bafkreiawvcmcwounww6dbzb2vlvvstqf7venmktd4tsgxkw4o4undmtipe"

	test := testUtils.TestCase{
		Description: "Test schema update, with branching schema",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// The second schema version will not be set as the active version, leaving the initial version active
				SetAsDefaultVersion: immutable.Some(false),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.SchemaPatch{
				// The third schema version will be set as the active version, going from version 1 to 3
				SetAsDefaultVersion: immutable.Some(true),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				// The email field is not queriable
				ExpectedError: `Cannot query field "email" on type "Users".`,
			},
			testUtils.GetSchema{
				// The second schema version is present in the system, with the email field
				VersionID: immutable.Some(schemaVersion2ID),
				ExpectedResults: []client.SchemaDescription{
					{
						Name:      "Users",
						VersionID: schemaVersion2ID,
						Root:      schemaVersion1ID,
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
							{
								Name: "email",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
			testUtils.Request{
				// The phone field is queriable
				Request: `query {
					Users {
						name
						phone
					}
				}`,
				Results: []map[string]any{},
			},
			testUtils.GetSchema{
				// The third schema version is present in the system, with the phone field
				VersionID: immutable.Some(schemaVersion3ID),
				ExpectedResults: []client.SchemaDescription{
					{
						Name:      "Users",
						VersionID: schemaVersion3ID,
						Root:      schemaVersion1ID,
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
							{
								Name: "phone",
								Kind: client.FieldKind_NILLABLE_STRING,
								Typ:  client.LWW_REGISTER,
							},
						},
					},
				},
			},
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
				},
				ExpectedResults: []client.CollectionDescription{
					{
						// The original collection version is present, it has no source and is inactive (has no name).
						ID:              1,
						SchemaVersionID: schemaVersion1ID,
					},
					{
						// The collection version for schema version 2 is present, it has the first collection as a source
						// and is inactive.
						ID:              2,
						SchemaVersionID: schemaVersion2ID,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: 1,
							},
						},
					},
					{
						// The collection version for schema version 3 is present and is active, it also has the first collection
						// as source.
						ID:              3,
						Name:            immutable.Some("Users"),
						SchemaVersionID: schemaVersion3ID,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: 1,
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdates_WithPatchOnBranchedSchema(t *testing.T) {
	schemaVersion1ID := "bafkreiht46o4lakri2py2zw57ed3pdeib6ud6ojlsomgjlrgwh53wl3q4a"
	schemaVersion2ID := "bafkreigdplzukezgpmjs45lw6kwzhtwge4xjzfgm6iodcd32d7kdageply"
	schemaVersion3ID := "bafkreiawvcmcwounww6dbzb2vlvvstqf7venmktd4tsgxkw4o4undmtipe"
	schemaVersion4ID := "bafkreidqp7ha7mfhwqpahevcpsn5etmi3soawyq76oytdxlyozvs6cgyui"

	test := testUtils.TestCase{
		Description: "Test schema update, with patch on branching schema",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// The second schema version will not be set as the active version, leaving the initial version active
				SetAsDefaultVersion: immutable.Some(false),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.SchemaPatch{
				// The third schema version will be set as the active version, going from version 1 to 3
				SetAsDefaultVersion: immutable.Some(true),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} }
					]
				`,
			},
			testUtils.SchemaPatch{
				// The fourth schema version will be set as the active version, going from version 3 to 4
				SetAsDefaultVersion: immutable.Some(true),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "discordName", "Kind": 11} }
					]
				`,
			},
			testUtils.Request{
				// The phone and discordName fields are queriable
				Request: `query {
					Users {
						name
						phone
						discordName
					}
				}`,
				Results: []map[string]any{},
			},
			testUtils.GetSchema{
				// The fourth schema version is present in the system, with the phone and discordName field
				VersionID: immutable.Some(schemaVersion4ID),
				ExpectedResults: []client.SchemaDescription{
					{
						Name:      "Users",
						VersionID: schemaVersion4ID,
						Root:      schemaVersion1ID,
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
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
				},
				ExpectedResults: []client.CollectionDescription{
					{
						// The original collection version is present, it has no source and is inactive (has no name).
						ID:              1,
						SchemaVersionID: schemaVersion1ID,
					},
					{
						// The collection version for schema version 2 is present, it has the first collection as a source
						// and is inactive.
						ID:              2,
						SchemaVersionID: schemaVersion2ID,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: 1,
							},
						},
					},
					{
						// The collection version for schema version 3 is present and inactive, it has the first collection
						// as source.
						ID:              3,
						SchemaVersionID: schemaVersion3ID,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: 1,
							},
						},
					},
					{
						// The collection version for schema version 4 is present and is active, it also has the third collection
						// as source.
						ID:              4,
						Name:            immutable.Some("Users"),
						SchemaVersionID: schemaVersion4ID,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: 3,
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdates_WithBranchingSchemaAndSetActiveSchemaToOtherBranch(t *testing.T) {
	schemaVersion1ID := "bafkreiht46o4lakri2py2zw57ed3pdeib6ud6ojlsomgjlrgwh53wl3q4a"
	schemaVersion2ID := "bafkreigdplzukezgpmjs45lw6kwzhtwge4xjzfgm6iodcd32d7kdageply"
	schemaVersion3ID := "bafkreiawvcmcwounww6dbzb2vlvvstqf7venmktd4tsgxkw4o4undmtipe"

	test := testUtils.TestCase{
		Description: "Test schema update, with branching schema toggling between branches",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// The second schema version will not be set as the active version, leaving the initial version active
				SetAsDefaultVersion: immutable.Some(false),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.SchemaPatch{
				// The third schema version will be set as the active version, going from version 1 to 3
				SetAsDefaultVersion: immutable.Some(true),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} }
					]
				`,
			},
			testUtils.SetActiveSchemaVersion{
				// Set the second schema version to be active
				SchemaVersionID: schemaVersion2ID,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				// The email field is queriable
				Results: []map[string]any{},
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						phone
					}
				}`,
				// The phone field is not queriable
				ExpectedError: `Cannot query field "phone" on type "Users".`,
			},
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
				},
				ExpectedResults: []client.CollectionDescription{
					{
						// The original collection version is present, it has no source and is inactive (has no name).
						ID:              1,
						SchemaVersionID: schemaVersion1ID,
					},
					{
						// The collection version for schema version 2 is present and is active, it has the first collection as a source
						ID:              2,
						Name:            immutable.Some("Users"),
						SchemaVersionID: schemaVersion2ID,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: 1,
							},
						},
					},
					{
						// The collection version for schema version 3 is present and is inactive, it also has the first collection
						// as source.
						ID:              3,
						SchemaVersionID: schemaVersion3ID,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: 1,
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdates_WithBranchingSchemaAndSetActiveSchemaToOtherBranchThenPatch(t *testing.T) {
	schemaVersion1ID := "bafkreiht46o4lakri2py2zw57ed3pdeib6ud6ojlsomgjlrgwh53wl3q4a"
	schemaVersion2ID := "bafkreigdplzukezgpmjs45lw6kwzhtwge4xjzfgm6iodcd32d7kdageply"
	schemaVersion3ID := "bafkreiawvcmcwounww6dbzb2vlvvstqf7venmktd4tsgxkw4o4undmtipe"
	schemaVersion4ID := "bafkreih5trmbzpjdgterha2amx2n6opgwlpvdyxfeyfi2uq7ncbodpl2cu"

	test := testUtils.TestCase{
		Description: "Test schema update, with branching schema toggling between branches then patch",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// The second schema version will not be set as the active version, leaving the initial version active
				SetAsDefaultVersion: immutable.Some(false),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.SchemaPatch{
				// The third schema version will be set as the active version, going from version 1 to 3
				SetAsDefaultVersion: immutable.Some(true),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} }
					]
				`,
			},
			testUtils.SetActiveSchemaVersion{
				// Set the second schema version to be active
				SchemaVersionID: schemaVersion2ID,
			},
			testUtils.SchemaPatch{
				// The fourth schema version will be set as the active version, going from version 2 to 4
				SetAsDefaultVersion: immutable.Some(true),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "discordName", "Kind": 11} }
					]
				`,
			},
			testUtils.Request{
				// The email and discordName fields are queriable
				Request: `query {
					Users {
						name
						email
						discordName
					}
				}`,
				Results: []map[string]any{},
			},
			testUtils.GetSchema{
				// The fourth schema version is present in the system, with the email and discordName field
				VersionID: immutable.Some(schemaVersion4ID),
				ExpectedResults: []client.SchemaDescription{
					{
						Name:      "Users",
						VersionID: schemaVersion4ID,
						Root:      schemaVersion1ID,
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
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
				},
				ExpectedResults: []client.CollectionDescription{
					{
						// The original collection version is present, it has no source and is inactive (has no name).
						ID:              1,
						SchemaVersionID: schemaVersion1ID,
					},
					{
						// The collection version for schema version 2 is present, it has the first collection as a source
						// and is inactive.
						ID:              2,
						SchemaVersionID: schemaVersion2ID,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: 1,
							},
						},
					},
					{
						// The collection version for schema version 3 is present and inactive, it has the first collection
						// as source.
						ID:              3,
						SchemaVersionID: schemaVersion3ID,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: 1,
							},
						},
					},
					{
						// The collection version for schema version 4 is present and is active, it also has the second collection
						// as source.
						ID:              4,
						Name:            immutable.Some("Users"),
						SchemaVersionID: schemaVersion4ID,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: 2,
							},
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdates_WithBranchingSchemaAndGetCollectionAtVersion(t *testing.T) {
	schemaVersion1ID := "bafkreiht46o4lakri2py2zw57ed3pdeib6ud6ojlsomgjlrgwh53wl3q4a"

	test := testUtils.TestCase{
		Description: `Test schema update, with branching schema toggling between branches and gets the 
collection at a specific version`,
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				// The second schema version will not be set as the active version, leaving the initial version active
				SetAsDefaultVersion: immutable.Some(true),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					SchemaVersionID: immutable.Some(schemaVersion1ID),
				},
				ExpectedResults: []client.CollectionDescription{
					{
						// The original collection version is present, it has no source and is inactive (has no name).
						ID:              1,
						SchemaVersionID: schemaVersion1ID,
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
