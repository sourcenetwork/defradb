// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_aac

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestAAC_GatesIndexDrop_AllowIfAuthorizedElseError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "admin acp correctly gates index drop operation, allow if authorized, otherwise error",
		Actions: []any{
			// Starting with ACC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},
			// Note: Doing setup steps after starting with aac enabled, otherwise the in-memory tests
			// will loose setup state when the restart happens (i.e. the restart that started aac).
			testUtils.SchemaUpdate{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type User {
						name: String @index
					}
				`,
			},
			// Note: Setup is now done, the test code that follows is what we want to assert.

			// We haven't authorized non-identities. So, this should error.
			testUtils.DropIndex{
				Identity:      testUtils.NoIdentity(),
				IndexName:     "User_name_ASC",
				ExpectedError: "not authorized to perform operation",
			},

			// Wrong user/identity will also not be authorized.
			testUtils.DropIndex{
				Identity:      testUtils.ClientIdentity(2),
				IndexName:     "User_name_ASC",
				ExpectedError: "not authorized to perform operation",
			},

			// This should work as the identity is authorized.
			testUtils.DropIndex{
				Identity:  testUtils.ClientIdentity(1),
				IndexName: "User_name_ASC",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
