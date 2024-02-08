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

func TestACP_AddPolicy_ExtraPermissions_ValidPolicyID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, extra permissions, still valid",

		Actions: []any{
			testUtils.AddPolicy{
				Creator: actor1Signature,

				Policy: `
                    description: a policy

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner
                          write:
                            expr: owner
                          extra:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor

                    actor:
                      name: actor
                `,

				ExpectedPolicyID: "9d518bb2d5aceb2c8f9b12b909eecd50276c1bd0250069875f265166e6030bb5",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_ExtraDuplicatePermissions_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, extra duplicate permissions, return error",

		Actions: []any{
			testUtils.AddPolicy{
				Creator: actor1Signature,

				Policy: `
                    description: a policy

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner
                          write:
                            expr: owner
                          write:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor

                    actor:
                      name: actor
                `,

				ExpectedError: "key \"write\" already set in map",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
