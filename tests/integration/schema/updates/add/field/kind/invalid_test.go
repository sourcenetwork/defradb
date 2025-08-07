// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kind

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesAddFieldKind15(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
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
func TestSchemaUpdatesAddFieldKind25(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
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
func TestSchemaUpdatesAddFieldKind198(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
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

func TestSchemaUpdatesAddFieldKindInvalid(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
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
