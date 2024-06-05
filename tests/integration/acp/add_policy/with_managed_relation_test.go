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

func TestACP_AddPolicy_WithRelationManagingOtherRelation_ValidPolicyID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, where a relation is managing another relation, valid policy id",
		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    name: a policy
                    description: a policy with admin relation managing reader relation

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader
                          write:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                          admin:
                            manages:
                              - reader
                            types:
                              - actor
                `,

				ExpectedPolicyID: "a42e109f1542da3fef5f8414621a09aa4805bf1ac9ff32ad9940bd2c488ee6cd",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
