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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateReplaceName_GivenExistingName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/Users/Name",
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

func TestColVersionUpdateReplaceName_GivenInactiveCollection_Errors(t *testing.T) {
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
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": "String"} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu/Name",
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
