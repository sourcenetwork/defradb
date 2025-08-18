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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateAddCollections_WithUndefinedID_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
				ExpectedError: "adding collections via patch is not supported",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateAddCollections_WithEmptyID_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
				ExpectedError: "adding collections via patch is not supported",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateAddCollections_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
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
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "add", "path": "/-", "value": {"Name": "Dogs"} }
					]
				`,
				ExpectedError: "adding collections via patch is not supported.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
