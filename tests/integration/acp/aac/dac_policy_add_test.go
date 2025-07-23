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

func TestAAC_GatesAddingDACPolicy_AllowIfAuthorizedElseError(t *testing.T) {
	policy := `
        name: Test Policy
        description: A Policy
        actor:
          name: actor
        resources:
          users:
            permissions:
              read:
                expr: owner + reader + updater + deleter
              update:
                expr: owner + updater
              delete:
                expr: owner + deleter
            relations:
              owner:
                types:
                  - actor
              reader:
                types:
                  - actor
              updater:
                types:
                  - actor
              deleter:
                types:
                  - actor
`

	test := testUtils.TestCase{
		Description: "admin acp correctly gates adding DAC policy operation, allow if authorized, otherwise error",
		Actions: []any{
			// Starting with ACC, so only authorized user(s) can perform operations from here on out.
			testUtils.Close{},
			testUtils.Start{
				Identity:  testUtils.ClientIdentity(1),
				EnableAAC: true,
			},

			// We haven't authorized non-identities. So, this should error.
			testUtils.AddDACPolicy{
				Identity:      testUtils.NoIdentity(),
				Policy:        policy,
				ExpectedError: "not authorized to perform operation",
			},

			// Wrong user/identity will also not be authorized.
			testUtils.AddDACPolicy{
				Identity:      testUtils.ClientIdentity(2),
				Policy:        policy,
				ExpectedError: "not authorized to perform operation",
			},

			// This should work as the identity is authorized.
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy:   policy,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
