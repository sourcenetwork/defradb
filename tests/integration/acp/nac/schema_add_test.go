// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_nac

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestNAC_GatesSchemaAdd_AllowIfAuthorizedElseError(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				// TODO: C binding test harness must be reworked to support this test
				// See: https://github.com/sourcenetwork/defradb/issues/3919
				testUtils.GoClientType,
				testUtils.CLIClientType,
				testUtils.HTTPClientType,
				testUtils.JSClientType,
			},
		),

		Actions: []any{
			// Starting with NAC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableNAC: true,
			},

			// We haven't authorized non-identities. So, this should error.
			&action.AddSchema{
				Identity: testUtils.NoIdentity(),
				Schema: `
					type Users {
						name: String
					}
				`,
				ExpectedError: "not authorized to perform operation",
			},

			// Wrong user/identity will also not be authorized.
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(2),
				Schema: `
					type Users {
						name: String
					}
				`,
				ExpectedError: "not authorized to perform operation",
			},

			// This should work as the identity is authorized.
			&action.AddSchema{
				Identity: testUtils.ClientIdentity(1),
				Schema: `
					type Users {
						name: String
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
