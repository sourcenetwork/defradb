// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateTestName(t *testing.T) {
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
							"op": "test",
							"path": "/bafyreihdbjfazsx5vq2tpzedqdktrjyn6lq22qle7el2s42b3q4zpxmwqq/Name",
							"value": "Users"
						}
					]
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColVersionUpdateTestName_Fails(t *testing.T) {
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
							"op": "test",
							"path": "/bafyreihdbjfazsx5vq2tpzedqdktrjyn6lq22qle7el2s42b3q4zpxmwqq/Name",
							"value": "Dogs"
						}
					]
				`,
				ExpectedError: "testing value /bafyreihdbjfazsx5vq2tpzedqdktrjyn6lq22qle7el2s42b3q4zpxmwqq/Name failed: test failed",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
