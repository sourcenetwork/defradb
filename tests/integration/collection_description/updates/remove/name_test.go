// Copyright 2024 Democratized Data Foundation
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

func TestColDescrUpdateRemoveName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "remove", "path": "/1/Name" }
					]
				`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
					}
				}`,
				// The Users collection has been deactivated and is no longer accessible
				ExpectedError: `Cannot query field "Users" on type "Query".`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
