// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_add_policy

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_AddPolicy_PermissionExprWithOwnerInTheEndWithMinus_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy with permission expr having owner in the end with minus, ValidID",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: reader - owner
                          write:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                `,

				ExpectedPolicyID: "fcb989d8bad149e3c4b22f8a69969760187b29ea1c796a3f9d2e16e32f493590",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Note: this and above test both result in different policy ids.
func TestACP_AddPolicy_PermissionExprWithOwnerInTheEndWithMinusNoSpace_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy with permission expr having owner in the end with minus no space, ValidID",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: reader-owner
                          write:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                `,

				ExpectedPolicyID: "50d8fbaf70a08c2c0e2bf0355a353a8bb06cc4d6e2f3ddbf71d91f9ef5aa49af",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
