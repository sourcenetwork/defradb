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

func TestColVersionUpdateReplaceFields_Errors(t *testing.T) {
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
							"op": "replace",
							"path": "/bafyreihuyovjl5ezgpud5xyqnouzsgx25x3ssrx3ncdv5p3guocc3laqna/Fields",
							"value": [{}]
						}
					]
				`,
				ExpectedError: "no type found for given name. Type: 0",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
func TestColVersionUpdateReplaceDefaultValue_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String @default(string: "Bob")
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu/Fields/1/DefaultValue",
							"value": "Alice"
						}
					]
				`,
				ExpectedError: "mutating an existing field is not supported",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
