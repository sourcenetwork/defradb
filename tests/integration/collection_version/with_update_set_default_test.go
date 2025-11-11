// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchema_WithUpdateAndSetDefaultVersionToEmptyString_Errors(t *testing.T) {
	test := testUtils.TestCase{
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
				VersionID:     "",
				ExpectedError: "schema version ID can't be empty",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchema_WithUpdateAndSetDefaultVersionToUnknownVersion_Errors(t *testing.T) {
	test := testUtils.TestCase{
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
				VersionID:     "does not exist",
				ExpectedError: "key not found",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchema_WithUpdateAndSetDefaultVersionToOriginal_NewFieldIsNotQueriable(t *testing.T) {
	test := testUtils.TestCase{
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: "bafyreiciz2hrrmt7ritk5gf5fyruw46v2tfhq5dc7qto4wgpzluben2smu",
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				// As the email field did not exist at this schema version, it will return a gql error
				ExpectedError: `Cannot query field "email" on type "Users".`,
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchema_WithUpdateAndSetDefaultVersionToNew_AllowsQueryingOfNewField(t *testing.T) {
	test := testUtils.TestCase{
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
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} },
						{ "op": "replace", "path": "/Users/IsActive", "value": false }
					]
				`,
			},
			testUtils.SetActiveCollectionVersion{
				VersionID: "bafyreigvzkfdc4y2ppvvpmmdw3t7kv4nd5dgfh5jfytef3kbzem6po55zu",
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
