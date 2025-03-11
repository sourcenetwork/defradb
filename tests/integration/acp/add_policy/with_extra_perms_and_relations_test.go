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
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
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

				ExpectedPolicyID: "d5128219b48e5f86ef54cbef8a5ea7ce2892e33e6df33519c7c6178cb285fc2d",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
