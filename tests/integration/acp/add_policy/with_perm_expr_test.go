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

				ExpectedPolicyID: "d74384d99b6732c3a6e0e47c7b75ea19553f643bcca416380530d8ad4e50e529",
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

				ExpectedPolicyID: "f6d5d6d8b0183230fcbdf06cfe14b611f782752d276006ad4622231eeaf60820",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
