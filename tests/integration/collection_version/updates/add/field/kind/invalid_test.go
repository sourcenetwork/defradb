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

package kind

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionVersionUpdatesAddFieldKind15(t *testing.T) {
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 15} }
					]
				`,
				ExpectedError: "no type found for given name. Type: 15",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// This test is currently the first unsupported value, if it becomes supported
// please update this test to be the newly lowest unsupported value.
func TestCollectionVersionUpdatesAddFieldKind25(t *testing.T) {
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 23} }
					]
				`,
				ExpectedError: "no type found for given name. Type: 23",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

// Tests a semi-random but hardcoded unsupported kind to try and protect against anything odd permitting
// high values.
func TestCollectionVersionUpdatesAddFieldKind198(t *testing.T) {
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": 198} }
					]
				`,
				ExpectedError: "no type found for given name. Type: 198",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestCollectionVersionUpdatesAddFieldKindInvalid(t *testing.T) {
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "foo", "Kind": "InvalidKind"} }
					]
				`,
				ExpectedError: "no type found for given name. Field: foo, Kind: InvalidKind",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
