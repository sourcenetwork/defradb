// Copyright 2025 Democratized Data Foundation
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

func TestColVersionUpdate_ReplaceFieldSizeContraint_ShouldError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						foo: [Int] @constraints(size: 2)
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/bafkreiegxruspmodnptoor6w5z6h42wjcao6souorcd2e5q3xtxxryxchu/Fields/1/Size",
							"value": 3
						}
					]
				`,
				ExpectedError: "collection fields cannot be mutated.",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
