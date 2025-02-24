// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package remove

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColDescrUpdate_FieldFieldSizeContraint_ShouldError(t *testing.T) {
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
						{ "op": "remove", "path": "/1/Fields/1/Size"}
					]
				`,
				ExpectedError: "collection fields cannot be mutated. CollectionID: 1",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
