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

func TestACP_AddPolicy_EmptyExpressionInPermission_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy with permission having empr expr, error",

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
                            expr:
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

				ExpectedError: "relation read: error parsing: expression needs: term",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_PermissionExprWithOwnerInTheEndWithInocorrectSymbol_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy with permission expr having owner in the end with incorrect symbol, error",

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
                            expr: reader ^ owner
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

				ExpectedError: "error parsing expression reader ^ owner: unknown token:",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_PermissionExprWithOwnerInTheEndWithInocorrectSymbolNoSpace_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy with permission expr having owner in the end with incorrect symbol with no space, error",

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
                            expr: reader^owner
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

				ExpectedError: "error parsing expression reader^owner: unknown token:",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
