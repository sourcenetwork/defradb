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

package move

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateMoveName(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {}
				`,
			},
			&action.PatchCollection{
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
