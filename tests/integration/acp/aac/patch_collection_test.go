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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestAAC_GatesPatchCollection_AllowIfAuthorizedElseError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "admin acp correctly gates patch collection operation, allow if authorized, otherwise error",
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
					type Users {}
				`,
			},
			testUtils.SchemaPatch{
				Identity: testUtils.ClientIdentity(1),
				Patch: `
					[
						{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "name", "Kind": "String"} }
					]
				`,
				SetAsDefaultVersion: immutable.Some(false),
			},
			// Note: Setup is now done, the test code that follows is what we want to assert.

			// We haven't authorized non-identities. So, this should error.
			testUtils.PatchCollection{
				Identity: testUtils.NoIdentity(),
				Patch: `
					[
						{
							"op": "copy",
							"from": "/bafkreia2jn5ecrhtvy4fravk6pm3wqiny46m7mqymvjkgat7xiqupgqoai/Name",
							"path": "/bafkreialnju2rez4t3quvpobf3463eai3lo64vdrdhdmunz7yy7sv3f5ce/Name"
						}
					]
				`,
				ExpectedError: "not authorized to perform operation",
			},

			// Wrong user/identity will also not be authorized.
			testUtils.PatchCollection{
				Identity: testUtils.ClientIdentity(2),
				Patch: `
					[
						{
							"op": "copy",
							"from": "/bafkreia2jn5ecrhtvy4fravk6pm3wqiny46m7mqymvjkgat7xiqupgqoai/Name",
							"path": "/bafkreialnju2rez4t3quvpobf3463eai3lo64vdrdhdmunz7yy7sv3f5ce/Name"
						}
					]
				`,
				ExpectedError: "not authorized to perform operation",
			},

			// This should work as the identity is authorized.
			testUtils.PatchCollection{
				Identity: testUtils.ClientIdentity(1),
				Patch: `
					[
						{
							"op": "copy",
							"from": "/bafkreia2jn5ecrhtvy4fravk6pm3wqiny46m7mqymvjkgat7xiqupgqoai/Name",
							"path": "/bafkreialnju2rez4t3quvpobf3463eai3lo64vdrdhdmunz7yy7sv3f5ce/Name"
						}
					]
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
