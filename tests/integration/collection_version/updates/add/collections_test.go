// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package add

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateAddCollections_WithUndefinedID_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/hgfgsagasga", "value": {"Name": "Dogs"} }
					]
				`,
				ExpectedError: "schema name can't be empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateAddCollections_WithEmptyID_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/hgfgsagasga", "value": {"VersionID": "", "Name": "Dogs"} }
					]
				`,
				ExpectedError: "schema name can't be empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateAddCollections_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "add",
							"path": "/hgfgsagasga",
							"value": {"VersionID": "hgfgsagasga", "Name": "Dogs"}
						}
					]
				`,
				ExpectedError: "adding collections via patch is not supported.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
func TestColVersionUpdateAddCollections_WithNoIndex_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/-", "value": {"Name": "Dogs"} }
					]
				`, // todo - doc properly
				// We get this error because we are marshalling into a map[uint32]CollectionVersion,
				// we will need to handle `-` when we allow adding collections via patches.
				ExpectedError: "schema name can't be empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
