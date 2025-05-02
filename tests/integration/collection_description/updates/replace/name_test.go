// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package replace

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColDescrUpdateReplaceName_GivenExistingName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafkreia3o3cetvcnnxyu5spucimoos77ifungfmacxdkva4zah2is3aooe/Name",
							"value": "Actors"
						}
					]
				`,
			},
			testUtils.GetCollections{
				ExpectedResults: []client.CollectionDescription{
					{
						Name:           immutable.Some("Actors"),
						IsMaterialized: true,
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
			testUtils.Request{
				Request: `query {
					Actors {
						name
					}
				}`,
				Results: map[string]any{
					"Actors": []map[string]any{
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

func TestColDescrUpdateReplaceName_GivenInactiveCollectionWithSameName_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": "String"} }
					]
				`,
				SetAsDefaultVersion: immutable.Some(false),
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafkreigtjpibdyrvmwvu7wbzatqpgavczrauj4huog2cvskwrgak6m7qgi/Name",
							"value": "Users"
						}
					]
				`,
				ExpectedError: "collection already exists. Name: Users",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdateReplaceName_GivenInactiveCollection_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": "String"} }
					]
				`,
				SetAsDefaultVersion: immutable.Some(false),
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafkreigtjpibdyrvmwvu7wbzatqpgavczrauj4huog2cvskwrgak6m7qgi/Name",
							"value": "Actors"
						}
					]
				`,
				// The params at the end of the error message is dependant on the order Go decides to iterate through
				// a map and so is not included in the test.
				ExpectedError: "multiple versions of same collection cannot be active",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdateReplaceName_RemoveExistingName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": "String"} }
					]
				`,
				SetAsDefaultVersion: immutable.Some(false),
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/bafkreia3o3cetvcnnxyu5spucimoos77ifungfmacxdkva4zah2is3aooe/Name"
						},
						{
							"op": "replace",
							"path": "/bafkreigtjpibdyrvmwvu7wbzatqpgavczrauj4huog2cvskwrgak6m7qgi/Name",
							"value": "Actors"
						}
					]
				`,
			},
			testUtils.GetCollections{
				FilterOptions: client.CollectionFetchOptions{
					IncludeInactive: immutable.Some(true),
				},
				ExpectedResults: []client.CollectionDescription{
					{
						IsMaterialized: true,
					},
					{
						Name:           immutable.Some("Actors"),
						IsMaterialized: true,
						Sources: []any{
							&client.CollectionSource{
								SourceCollectionID: "bafkreia3o3cetvcnnxyu5spucimoos77ifungfmacxdkva4zah2is3aooe",
							},
						},
					},
				},
			},
			testUtils.Request{
				Request: `query {
					Actors {
						name
					}
				}`,
				Results: map[string]any{
					"Actors": []map[string]any{
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
