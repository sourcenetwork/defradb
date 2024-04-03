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

func TestACP_AddPolicy_MissingRequiredOwnerRelation_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, missing requred owner relation, should return error",

		Actions: []any{
			testUtils.AddPolicy{
				Creator: actor1Signature,

				Policy: `
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          write:
                            expr: reader
                          read:
                            expr: reader

                        relations:
                          reader:
                            types:
                              - actor
                `,

				ExpectedError: "resource users: resource missing owner relation: invalid policy",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_DuplicateOwnerRelation_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, duplicate required owner relations, return error",

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

                        relations:
                          owner:
                            types:
                              - actor
                          owner:
                            types:
                              - actor

                    actor:
                      name: actor
                `,

				ExpectedError: "key \"owner\" already set in map",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
