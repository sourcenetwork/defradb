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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateReplaceFields_Errors(t *testing.T) {
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
							"op": "replace",
							"path": "/bafkreia2jn5ecrhtvy4fravk6pm3wqiny46m7mqymvjkgat7xiqupgqoai/Fields",
							"value": [{}]
						}
					]
				`,
				ExpectedError: "collection fields cannot be mutated.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
func TestColVersionUpdateReplaceDefaultValue_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
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
							"path": "/bafkreia3o3cetvcnnxyu5spucimoos77ifungfmacxdkva4zah2is3aooe/Fields/1/DefaultValue",
							"value": "Alice"
						}
					]
				`,
				ExpectedError: "collection fields cannot be mutated.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
