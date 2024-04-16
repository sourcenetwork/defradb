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

func TestACP_AddPolicy_EmptyPolicyData_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, adding empty policy, return error",

		Actions: []any{
			testUtils.AddPolicy{
				Creator: actor1Identity,

				Policy: "",

				ExpectedError: "policy data can not be empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_EmptyPolicyCreator_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, adding policy, with empty creator, return error",

		Actions: []any{
			testUtils.AddPolicy{
				Creator: "",

				Policy: `
                    description: a basic policy that satisfies minimum DPI requirements

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner
                          write:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor

                `,

				ExpectedError: "policy creator can not be empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_EmptyCreatorAndPolicyArgs_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, adding policy, with empty policy and empty creator, return error",

		Actions: []any{
			testUtils.AddPolicy{
				Creator: "",

				Policy: "",

				ExpectedError: "policy creator can not be empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
