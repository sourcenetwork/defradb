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
				ExpectedError: "collection name cannot be mutated.",
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
				ExpectedError: "collection name cannot be mutated.",
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
				ExpectedError: "collection name can't be empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
