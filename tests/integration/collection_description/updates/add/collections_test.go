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

func TestColDescrUpdateAddCollections_WithUndefinedID_Errors(t *testing.T) {
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
						{ "op": "add", "path": "/2", "value": {"Name": "Dogs"} }
					]
				`,
				ExpectedError: "collection ID cannot be zero",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdateAddCollections_Errors(t *testing.T) {
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
						{ "op": "add", "path": "/2", "value": {"ID": 2, "Name": "Dogs"} }
					]
				`,
				ExpectedError: "adding collections via patch is not supported. ID: 2",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdateAddCollections_WithNoIndex_Errors(t *testing.T) {
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
				`,
				// We get this error because we are marshalling into a map[uint32]CollectionDescription,
				// we will need to handle `-` when we allow adding collections via patches.
				ExpectedError: "json: cannot unmarshal number - into Go value of type uint32",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
