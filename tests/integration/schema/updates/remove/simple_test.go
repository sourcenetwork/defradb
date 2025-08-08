// Copyright 2023 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchemaUpdatesRemoveCollectionNameErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove collection name",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "remove", "path": "/Users/Name" }
					]
				`,
				ExpectedError: "collection name can't be empty",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesRemoveVersionIDErrors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "remove", "path": "/Users/VersionID" }
					]
				`,
				ExpectedError: "invalid cid: cid too short. VersionID:",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaUpdatesRemoveSchemaNameErrors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, remove schema name",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "remove", "path": "/Users/Name" }
					]
				`,
				ExpectedError: "collection name can't be empty",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
