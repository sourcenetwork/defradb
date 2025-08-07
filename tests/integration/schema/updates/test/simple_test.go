// Copyright 2023 Democratized Data Foundation
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

func TestSchemaUpdatesTestCollectionNameErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, test collection name",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "test", "path": "/Users/Name", "value": "Book" }
					]
				`,
				ExpectedError: "failed: test failed",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesTestCollectionNamePasses(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, test collection name passes",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "test", "path": "/Users/Name", "value": "Users" }
					]
				`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
