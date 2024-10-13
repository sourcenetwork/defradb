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
				Identity: testUtils.UserIdentity(1),

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

				ExpectedPolicyID: "af2a2eaa2d6701262ea60665487c87e3d41ab727194e1ea18ec16348149a02cc",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
