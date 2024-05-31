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

func TestACP_AddPolicy_UnusedRelation_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, unused relation in permissions",

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
                            expr: owner
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

				ExpectedPolicyID: "a6d42bfedff5db1feca0313793e4f9540851e3feaefffaebc98a1ee5bb140e45",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
