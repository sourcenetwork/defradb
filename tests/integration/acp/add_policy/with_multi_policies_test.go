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

func TestACP_AddPolicy_AddMultipleDifferentPolicies_ValidPolicyIDs(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add multiple different policies",

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
                          read:
                            expr: owner
                          write:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor

                `,

				ExpectedPolicyID: "2eb8b503c9fc0b7c1f7b04d68a6faa0f82a299db0ae02fed68f9897612439cb6",
			},

			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    name: a policy
                    description: another policy

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

				ExpectedPolicyID: "6b766a9aafabf0bf65102f73b7cd81963b65e1fd87ce763f386cc685147325ee",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_AddMultipleDifferentPoliciesInDifferentFmts_ValidPolicyIDs(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add multiple different policies in different formats",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    {
                      "name": "test",
                      "description": "a policy",
                      "actor": {
                        "name": "actor"
                      },
                      "resources": {
                        "users": {
                          "permissions": {
                            "read": {
                              "expr": "owner"
                            },
                            "write": {
                              "expr": "owner"
                            }
                          },
                          "relations": {
                            "owner": {
                              "types": [
                                "actor"
                              ]
                            }
                          }
                        }
                      }
                    }
                `,

				ExpectedPolicyID: "66f3e364004a181e9b129f65dea317322d2285226e926d7e8cdfd644954e4262",
			},

			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    name: test
                    description: another policy

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

				ExpectedPolicyID: "7bcb558ef8dac6b744a11ea144a61a756ea38475554097ac04612037c36ffe52",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_AddDuplicatePolicyByOtherCreator_ValidPolicyIDs(t *testing.T) {
	const policyUsedByBoth string = `
        name: test
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
    `

	test := testUtils.TestCase{

		Description: "Test acp, add duplicate policies by different actors, valid",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: policyUsedByBoth,

				ExpectedPolicyID: "66f3e364004a181e9b129f65dea317322d2285226e926d7e8cdfd644954e4262",
			},

			testUtils.AddPolicy{
				Identity: actor2Identity,

				Policy: policyUsedByBoth,

				ExpectedPolicyID: "ec02815cb630850678bda5e2d75cfacebc96f5610e32a602f7bfc414e21474ad",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_AddMultipleDuplicatePolicies_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add duplicate policies, error",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    name: test
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

                `,

				ExpectedPolicyID: "66f3e364004a181e9b129f65dea317322d2285226e926d7e8cdfd644954e4262",
			},

			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    name: test
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

                `,

				ExpectedPolicyID: "ec02815cb630850678bda5e2d75cfacebc96f5610e32a602f7bfc414e21474ad",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_AddMultipleDuplicatePoliciesDifferentFmts_ProducesDifferentIDs(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add duplicate policies different formats, error",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    name: test
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
                `,

				ExpectedPolicyID: "66f3e364004a181e9b129f65dea317322d2285226e926d7e8cdfd644954e4262",
			},

			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                   {
                     "name": "test",
                     "description": "a policy",
                     "actor": {
                       "name": "actor"
                     },
                     "resources": {
                       "users": {
                         "permissions": {
                           "read": {
                             "expr": "owner"
                           },
                           "write": {
                             "expr": "owner"
                           }
                         },
                         "relations": {
                           "owner": {
                             "types": [
                               "actor"
                             ]
                           }
                         }
                       }
                     }
                   }
               `,

				ExpectedPolicyID: "ec02815cb630850678bda5e2d75cfacebc96f5610e32a602f7bfc414e21474ad",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
