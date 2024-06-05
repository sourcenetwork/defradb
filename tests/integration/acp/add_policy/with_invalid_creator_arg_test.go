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

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

func TestACP_AddPolicy_InvalidCreatorIdentityWithValidPolicy_Error(t *testing.T) {
	test := testUtils.TestCase{
		// Using an invalid creator is not possible with other client
		// types since the token authentication will fail
		SupportedClientTypes: immutable.Some([]testUtils.ClientType{
			testUtils.GoClientType,
		}),

		Description: "Test acp, adding policy, with invalid creator, with valid policy, return error",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: immutable.Some(acpIdentity.Identity{Address: "invalid"}),

				Policy: `
                    name: a policy
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

	//TODO-ACP: https://github.com/sourcenetwork/defradb/issues/2357
	testUtils.AssertPanic(t, func() { testUtils.ExecuteTestCase(t, test) })
}

func TestACP_AddPolicy_InvalidCreatorIdentityWithEmptyPolicy_Error(t *testing.T) {
	test := testUtils.TestCase{
		// Using an invalid creator is not possible with other client
		// types since the token authentication will fail
		SupportedClientTypes: immutable.Some([]testUtils.ClientType{
			testUtils.GoClientType,
		}),

		Description: "Test acp, adding policy, with invalid creator, with empty policy, return error",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: immutable.Some(acpIdentity.Identity{Address: "invalid"}),

				Policy: "",

				ExpectedError: "policy data can not be empty",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
