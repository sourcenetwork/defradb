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
				IsYAML: true,

				Creator: "cosmos1zzg43wdrhmmk89z3pmejwete2kkd4a3vn7w969",

				Policy: `
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

				ExpectedPolicyID: "53980e762616fcffbe76307995895e862f87ef3f21d509325d1dc772a770b001",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
