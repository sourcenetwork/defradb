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

func TestColDescrUpdateReplaceIsBranchable_UpdatingFromTrueToFalse_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @branchable {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "replace", "path": "/1/IsBranchable", "value": false }
					]
				`,
				ExpectedError: "mutating IsBranchable is not supported. Collection: User",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdateReplaceIsBranchable_UpdatingFromFalseToTrue_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User @branchable(if: false) {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "replace", "path": "/1/IsBranchable", "value": true }
					]
				`,
				ExpectedError: "mutating IsBranchable is not supported. Collection: User",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
