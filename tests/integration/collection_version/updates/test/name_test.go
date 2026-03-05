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

package test

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateTestNameByVersionID(t *testing.T) {
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
							"op": "test",
							"path": "/bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna/Name",
							"value": "Users"
						}
					]
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateTestNameByVersionID_Fails(t *testing.T) {
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
							"op": "test",
							"path": "/bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna/Name",
							"value": "Dogs"
						}
					]
				`,
				ExpectedError: "testing value /bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna/Name failed: test failed",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateTestName(t *testing.T) {
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
						{ "op": "test", "path": "/Users/Name", "value": "Users" }
					]
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateTestName_Fails(t *testing.T) {
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
						{ "op": "test", "path": "/Users/Name", "value": "Book" }
					]
				`,
				ExpectedError: "failed: test failed",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
