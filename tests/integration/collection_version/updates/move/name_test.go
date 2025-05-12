// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package move

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateMoveName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {}
				`,
			},
			testUtils.PatchCollection{
				// Make the second collection the active one by moving its name from the first to the second
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} },
						{
							"op": "move",
							"from": "/Users/Name",
							"path": "/Users/Fields/1/Name"
						},
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
				ExpectedError: "collection name can't be empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
