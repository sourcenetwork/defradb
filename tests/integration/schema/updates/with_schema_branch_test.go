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
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdates_WithBranchingSchema(t *testing.T) {
	schemaVersion1ID := "bafyreigsld6ten2pppcu2tgkbexqwdndckp6zt2vfjhuuheykqkgpmwk7i"
	schemaVersion2ID := "bafyreiav27gqgcudly2dige7m72giaaucv4fr2ko225rnvfyyauvpmho6a"
	schemaVersion3ID := "bafyreih3tqiftewsy5mb2wpsuuinkfptxlrh36c7a36bkehnmyn6lodvau"

	test := testUtils.TestCase{
		Description: "Test schema update, with branching schema",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				// The second schema version will not be set as the active version, leaving the initial version active
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			testUtils.PatchCollection{
				// The third schema version will be set as the active version, going from version 1 to 3
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
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
			testUtils.GetCollections{
				// The second schema version is present in the system, with the email field
				FilterOptions: client.CollectionFetchOptions{
					VersionID: immutable.Some(schemaVersion2ID),
				},
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						VersionID:      schemaVersion2ID,
						CollectionID:   schemaVersion1ID,
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
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
			testUtils.GetCollections{
				// The third schema version is present in the system, with the phone field
				FilterOptions: client.CollectionFetchOptions{
					VersionID: immutable.Some(schemaVersion3ID),
				},
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						VersionID:      schemaVersion3ID,
						CollectionID:   schemaVersion1ID,
						IsActive:       true,
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
				ExpectedResults: []client.CollectionVersion{
					{
						// The collection version for schema version 2 is present, it has the first collection as a source
						// and is inactive.
						Name:           "Users",
						VersionID:      schemaVersion2ID,
						IsMaterialized: true,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: schemaVersion1ID,
							},
						},
					},
					{
						// The original collection version is present, it has no source and is inactive (has no name).
						VersionID:      schemaVersion1ID,
						IsMaterialized: true,
						Name:           "Users",
					},
					{
						// The collection version for schema version 3 is present and is active, it also has the first collection
						// as source.
						Name:           "Users",
						VersionID:      schemaVersion3ID,
						IsMaterialized: true,
						IsActive:       true,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: schemaVersion1ID,
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
	schemaVersion1ID := "bafyreigsld6ten2pppcu2tgkbexqwdndckp6zt2vfjhuuheykqkgpmwk7i"
	schemaVersion2ID := "bafyreiav27gqgcudly2dige7m72giaaucv4fr2ko225rnvfyyauvpmho6a"
	schemaVersion3ID := "bafyreih3tqiftewsy5mb2wpsuuinkfptxlrh36c7a36bkehnmyn6lodvau"
	schemaVersion4ID := "bafyreih3jgjlud3h3xnnjysgtvfegcn2xdixlavwzbw7s6as65atwfhtvi"

	test := testUtils.TestCase{
		Description: "Test schema update, with patch on branching schema",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				// The second schema version will not be set as the active version, leaving the initial version active
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			testUtils.PatchCollection{
				// The third schema version will be set as the active version, going from version 1 to 3
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
					]
				`,
			},
			testUtils.PatchCollection{
				// The fourth schema version will be set as the active version, going from version 3 to 4
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "discordName", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
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
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
			testUtils.GetCollections{
				// The fourth schema version is present in the system, with the phone and discordName field
				FilterOptions: client.CollectionFetchOptions{
					VersionID: immutable.Some(schemaVersion4ID),
				},
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						VersionID:      schemaVersion4ID,
						CollectionID:   schemaVersion1ID,
						IsActive:       true,
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
				ExpectedResults: []client.CollectionVersion{
					{
						// The collection version for schema version 2 is present, it has the first collection as a source
						// and is inactive.
						Name:           "Users",
						VersionID:      schemaVersion2ID,
						IsMaterialized: true,
						IsActive:       false,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: schemaVersion1ID,
							},
						},
					},
					{
						// The original collection version is present, it has no source and is inactive
						Name:           "Users",
						VersionID:      schemaVersion1ID,
						IsMaterialized: true,
						IsActive:       false,
					},
					{
						// The collection version for schema version 4 is present and is active, it also has the third collection
						// as source.
						Name:           "Users",
						VersionID:      schemaVersion4ID,
						IsMaterialized: true,
						IsActive:       true,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: schemaVersion3ID,
							},
						},
					},
					{
						// The collection version for schema version 3 is present and inactive, it has the first collection
						// as source.
						Name:           "Users",
						VersionID:      schemaVersion3ID,
						IsMaterialized: true,
						IsActive:       false,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: schemaVersion1ID,
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
	schemaVersion1ID := "bafyreigsld6ten2pppcu2tgkbexqwdndckp6zt2vfjhuuheykqkgpmwk7i"
	schemaVersion2ID := "bafyreiav27gqgcudly2dige7m72giaaucv4fr2ko225rnvfyyauvpmho6a"
	schemaVersion3ID := "bafyreih3tqiftewsy5mb2wpsuuinkfptxlrh36c7a36bkehnmyn6lodvau"

	test := testUtils.TestCase{
		Description: "Test schema update, with branching schema toggling between branches",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				// The second schema version will not be set as the active version, leaving the initial version active
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			testUtils.PatchCollection{
				// The third schema version will be set as the active version, going from version 1 to 3
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
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
				Results: map[string]any{
					"Users": []map[string]any{},
				},
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
				ExpectedResults: []client.CollectionVersion{
					{
						// The collection version for schema version 2 is present and is active, it has the first collection as a source
						Name:           "Users",
						VersionID:      schemaVersion2ID,
						IsMaterialized: true,
						IsActive:       true,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: schemaVersion1ID,
							},
						},
					},
					{
						// The original collection version is present, it has no source and is inactive.
						Name:           "Users",
						VersionID:      schemaVersion1ID,
						IsMaterialized: true,
						IsActive:       false,
					},
					{
						// The collection version for schema version 3 is present and is inactive, it also has the first collection
						// as source.
						Name:           "Users",
						VersionID:      schemaVersion3ID,
						IsMaterialized: true,
						IsActive:       false,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: schemaVersion1ID,
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
	schemaVersion1ID := "bafyreigsld6ten2pppcu2tgkbexqwdndckp6zt2vfjhuuheykqkgpmwk7i"
	schemaVersion2ID := "bafyreiav27gqgcudly2dige7m72giaaucv4fr2ko225rnvfyyauvpmho6a"
	schemaVersion3ID := "bafyreih3tqiftewsy5mb2wpsuuinkfptxlrh36c7a36bkehnmyn6lodvau"
	schemaVersion4ID := "bafyreih3jgjlud3h3xnnjysgtvfegcn2xdixlavwzbw7s6as65atwfhtvi"

	test := testUtils.TestCase{
		Description: "Test schema update, with branching schema toggling between branches then patch",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				// The second schema version will not be set as the active version, leaving the initial version active
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			testUtils.PatchCollection{
				// The third schema version will be set as the active version, going from version 1 to 3
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "phone", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
				// Set the second schema version to be active
				SchemaVersionID: schemaVersion2ID,
			},
			testUtils.PatchCollection{
				// The fourth schema version will be set as the active version, going from version 2 to 4
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "discordName", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
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
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
			testUtils.GetCollections{
				// The fourth schema version is present in the system, with the email and discordName field
				FilterOptions: client.CollectionFetchOptions{
					VersionID: immutable.Some(schemaVersion4ID),
				},
				ExpectedResults: []client.CollectionVersion{
					{
						Name:           "Users",
						VersionID:      schemaVersion4ID,
						CollectionID:   schemaVersion1ID,
						IsActive:       true,
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
				ExpectedResults: []client.CollectionVersion{
					{
						// The collection version for schema version 2 is present, it has the first collection as a source
						// and is inactive.
						Name:           "Users",
						VersionID:      schemaVersion2ID,
						IsMaterialized: true,
						IsActive:       false,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: schemaVersion1ID,
							},
						},
					},
					{
						// The original collection version is present, it has no source and is inactive.
						Name:           "Users",
						VersionID:      schemaVersion1ID,
						IsMaterialized: true,
						IsActive:       false,
					},
					{
						// The collection version for schema version 4 is present and is active, it also has the second collection
						// as source.
						Name:           "Users",
						VersionID:      schemaVersion4ID,
						IsMaterialized: true,
						IsActive:       true,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: schemaVersion2ID,
							},
						},
					},
					{
						// The collection version for schema version 3 is present and inactive, it has the first collection
						// as source.
						Name:           "Users",
						VersionID:      schemaVersion3ID,
						IsMaterialized: true,
						IsActive:       false,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: schemaVersion1ID,
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
	schemaVersion1ID := "bafyreigsld6ten2pppcu2tgkbexqwdndckp6zt2vfjhuuheykqkgpmwk7i"

	test := testUtils.TestCase{
		Description: `Test schema update, with branching schema toggling between branches and gets the 
collection at a specific version`,
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				// The second schema version will not be set as the active version, leaving the initial version active
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": true }
					]
				`,
			},
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					VersionID: immutable.Some(schemaVersion1ID),
				},
				ExpectedResults: []client.CollectionVersion{
					{
						// The original collection version is present, it has no source and is inactive.
						Name:           "Users",
						VersionID:      schemaVersion1ID,
						IsMaterialized: true,
						IsActive:       false,
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
