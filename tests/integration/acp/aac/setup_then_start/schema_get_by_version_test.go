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

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestAAC_GatesSchemaGetByVersionPreSetup_AllowIfAuthorizedElseError(t *testing.T) {
	schemaVersionID := "bafkreia2jn5ecrhtvy4fravk6pm3wqiny46m7mqymvjkgat7xiqupgqoai"

	test := testUtils.TestCase{
		Description: "admin acp correctly gates schema get by version operation (setup before aac), allow if authorized, otherwise error",
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
					type Users {}
				`,
			},

			// Starting with ACC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// We haven't authorized non-identities. So, this should error.
			testUtils.GetSchema{
				Identity:      testUtils.NoIdentity(),
				VersionID:     immutable.Some(schemaVersionID),
				ExpectedError: "not authorized to perform operation",
			},

			// Wrong user/identity will also not be authorized.
			testUtils.GetSchema{
				Identity:      testUtils.ClientIdentity(2),
				VersionID:     immutable.Some(schemaVersionID),
				ExpectedError: "not authorized to perform operation",
			},

			// This should work as the identity is authorized.
			testUtils.GetSchema{
				Identity:  testUtils.ClientIdentity(1),
				VersionID: immutable.Some(schemaVersionID),
				ExpectedResults: []client.SchemaDescription{
					{
						Name:      "Users",
						Root:      schemaVersionID,
						VersionID: schemaVersionID,
						Fields: []client.SchemaFieldDescription{
							{
								Name: "_docID",
								Kind: client.FieldKind_DocID,
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
