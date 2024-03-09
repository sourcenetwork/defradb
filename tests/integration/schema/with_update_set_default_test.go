// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestSchema_WithUpdateAndSetDefaultVersionToEmptyString_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, set default version to empty string",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.SetActiveSchemaVersion{
				SchemaVersionID: "",
				ExpectedError:   "schema version ID can't be empty",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchema_WithUpdateAndSetDefaultVersionToUnknownVersion_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, set default version to invalid string",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
			},
			testUtils.SetActiveSchemaVersion{
				SchemaVersionID: "does not exist",
				ExpectedError:   "datastore: key not found",
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestSchema_WithUpdateAndSetDefaultVersionToOriginal_NewFieldIsNotQueriable(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test schema update, set default version to original schema version",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
				SetAsDefaultVersion: immutable.Some(false),
			},
			testUtils.SetActiveSchemaVersion{
				SchemaVersionID: "bafkreiebcgze3rs6j3g7gu65dwskdg5fn3qby5c6nqffhbdkcy2l5bbvp4",
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
		Description: "Test schema update, set default version to new schema version",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.SchemaPatch{
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": 11} }
					]
				`,
				SetAsDefaultVersion: immutable.Some(false),
			},
			testUtils.SetActiveSchemaVersion{
				SchemaVersionID: "bafkreidn4f3i52756wevi3sfpbqzijgy6v24zh565pmvtmpqr4ou52v2q4",
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				Results: []map[string]any{},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
