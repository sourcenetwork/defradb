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

func TestACP_AddPolicy_ExtraPermissionsAndExtraRelations_ValidPolicyID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, extra permissions and relations, still valid",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    name: a policy
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          write:
                            expr: owner
                          read:
                            expr: owner + reader
                          extra:
                            expr: joker

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                          joker:
                            types:
                              - actor
                `,

				ExpectedPolicyID: "f29c97dca930c9e93f7ef9e2139c63939c573af96c95af5cb9392861a0111b13",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
