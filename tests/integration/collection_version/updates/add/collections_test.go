// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package add

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateAddCollections_WithUndefinedID_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
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
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
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
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
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
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
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
