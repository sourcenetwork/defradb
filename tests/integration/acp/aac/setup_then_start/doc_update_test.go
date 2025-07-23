// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_aac_setup_then_start

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestAAC_GatesDocUpdatePreSetup_AllowIfAuthorizedElseError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "admin acp correctly gates doc update operation (setup before aac), allow if authorized, otherwise error",
		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since the setup steps will be done before
				// the node is re-started with aac enabled (if it's in-memory it will loose setup state).
				testUtils.BadgerFileType,
			},
		),
		Actions: []any{
			// Note: Since this is not an in-memory test, we can do the setup steps before aac is enabled.
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Shahzad",
					"age": 48
				}`,
			},

			// Starting with ACC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// We haven't authorized non-identities. So, this should error.
			testUtils.UpdateDoc{
				Identity:      testUtils.NoIdentity(),
				CollectionID:  0,
				DocID:         0,
				Doc:           `{ "name": "Lone" }`,
				ExpectedError: "not authorized to perform operation",
			},
			testUtils.Request{ // Should not be updated
				Identity: testUtils.ClientIdentity(1),
				Request:  `query{ User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Shahzad"}},
				},
			},

			// Wrong user/identity will also not be authorized.
			testUtils.UpdateDoc{
				Identity:      testUtils.ClientIdentity(2),
				CollectionID:  0,
				DocID:         0,
				Doc:           `{ "name": "Lone" }`,
				ExpectedError: "not authorized to perform operation",
			},
			testUtils.Request{ // Should not be updated
				Identity: testUtils.ClientIdentity(1),
				Request:  `query{ User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Shahzad"}},
				},
			},

			// This should work as the identity is authorized.
			testUtils.UpdateDoc{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				DocID:        0,
				Doc:          `{ "name": "Lone" }`,
			},
			testUtils.Request{ // Should now be updated
				Identity: testUtils.ClientIdentity(1),
				Request:  `query{ User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Lone"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestAAC_GatesDocUpdateWithFilterPreSetup_AllowIfAuthorizedElseError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "admin acp correctly gates doc update with filter operation (setup before aac), allow if authorized, otherwise error",
		SupportedDatabaseTypes: immutable.Some(
			[]testUtils.DatabaseType{
				// This test only supports file type databases since the setup steps will be done before
				// the node is re-started with aac enabled (if it's in-memory it will loose setup state).
				testUtils.BadgerFileType,
			},
		),
		Actions: []any{
			// Note: Since this is not an in-memory test, we can do the setup steps before aac is enabled.
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						age: Int 
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Shahzad",
					"age": 48
				}`,
			},

			// Starting with ACC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// We haven't authorized non-identities. So, this should error.
			testUtils.UpdateWithFilter{
				Identity:      testUtils.NoIdentity(),
				CollectionID:  0,
				Filter:        `{name: {_eq: "Shahzad"}}`,
				Updater:       `{"name": "Lone"}`,
				ExpectedError: "not authorized to perform operation",
			},
			testUtils.Request{ // Should not be updated
				Identity: testUtils.ClientIdentity(1),
				Request:  `query{ User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Shahzad"}},
				},
			},

			// Wrong user/identity will also not be authorized.
			testUtils.UpdateWithFilter{
				Identity:      testUtils.ClientIdentity(2),
				CollectionID:  0,
				Filter:        `{name: {_eq: "Shahzad"}}`,
				Updater:       `{"name": "Lone"}`,
				ExpectedError: "not authorized to perform operation",
			},
			testUtils.Request{ // Should not be updated
				Identity: testUtils.ClientIdentity(1),
				Request:  `query{ User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Shahzad"}},
				},
			},

			// This should work as the identity is authorized.
			testUtils.UpdateWithFilter{
				Identity:     testUtils.ClientIdentity(1),
				CollectionID: 0,
				Filter:       `{name: {_eq: "Shahzad"}}`,
				Updater:      `{"name": "Lone"}`,
			},
			testUtils.Request{ // Should now be updated
				Identity: testUtils.ClientIdentity(1),
				Request:  `query{ User { name } }`,
				Results: map[string]any{
					"User": []map[string]any{{"name": "Lone"}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
