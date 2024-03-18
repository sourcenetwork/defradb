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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColDescrUpdateTestName(t *testing.T) {
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
						{ "op": "test", "path": "/1/Name", "value": "Users" }
					]
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestColDescrUpdateTestName_Fails(t *testing.T) {
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
						{ "op": "test", "path": "/1/Name", "value": "Dogs" }
					]
				`,
				ExpectedError: "testing value /1/Name failed: test failed",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
