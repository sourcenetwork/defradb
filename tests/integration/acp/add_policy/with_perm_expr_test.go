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
				Identity: testUtils.ClientIdentity(1),

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

				ExpectedPolicyID: "0d428b097234d2da7734f8464890ec3f7ad346ac014ecfbb150dc5d6bf457f8b",
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
				Identity: testUtils.ClientIdentity(1),

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

				ExpectedPolicyID: "cce1b5c107c8507c56b6209f19f8b2d2fe414591ec0a4061f6c397b1cce873c6",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
