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
				IsYAML: true,

				Creator: actor1Signature,

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

                `,

				ExpectedPolicyID: "dfe202ffb4f0fe9b46157c313213a3839e08a6f0a7c3aba55e4724cb49ffde8a",
			},

			testUtils.AddPolicy{
				IsYAML: true,

				Creator: actor1Signature,

				Policy: `
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

				ExpectedPolicyID: "53980e762616fcffbe76307995895e862f87ef3f21d509325d1dc772a770b001",
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
				IsYAML: false,

				Creator: actor1Signature,

				Policy: `
                    {
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

				ExpectedPolicyID: "dfe202ffb4f0fe9b46157c313213a3839e08a6f0a7c3aba55e4724cb49ffde8a",
			},

			testUtils.AddPolicy{
				IsYAML: true,

				Creator: actor1Signature,

				Policy: `
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

				ExpectedPolicyID: "53980e762616fcffbe76307995895e862f87ef3f21d509325d1dc772a770b001",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_AddDuplicatePolicyByOtherCreator_ValidPolicyIDs(t *testing.T) {
	const policyUsedByBoth string = `
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
				IsYAML: true,

				Creator: actor1Signature,

				Policy: policyUsedByBoth,

				ExpectedPolicyID: "dfe202ffb4f0fe9b46157c313213a3839e08a6f0a7c3aba55e4724cb49ffde8a",
			},

			testUtils.AddPolicy{
				IsYAML: true,

				Creator: actor2Signature,

				Policy: policyUsedByBoth,

				ExpectedPolicyID: "551c57323f33decfdc23312e5e1036e3ab85d2414e962814dab9101619dd9ff9",
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
				IsYAML: true,

				Creator: actor1Signature,

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

                `,

				ExpectedPolicyID: "dfe202ffb4f0fe9b46157c313213a3839e08a6f0a7c3aba55e4724cb49ffde8a",
			},

			testUtils.AddPolicy{
				IsYAML: true,

				Creator: actor1Signature,

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

                `,

				ExpectedError: "policy dfe202ffb4f0fe9b46157c313213a3839e08a6f0a7c3aba55e4724cb49ffde8a: policy exists",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_AddMultipleDuplicatePoliciesDifferentFmts_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add duplicate policies different formats, error",

		Actions: []any{
			testUtils.AddPolicy{
				IsYAML: true,

				Creator: actor1Signature,

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
                `,

				ExpectedPolicyID: "dfe202ffb4f0fe9b46157c313213a3839e08a6f0a7c3aba55e4724cb49ffde8a",
			},

			testUtils.AddPolicy{
				IsYAML: false,

				Creator: actor1Signature,

				Policy: `
                   {
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

				ExpectedError: "policy dfe202ffb4f0fe9b46157c313213a3839e08a6f0a7c3aba55e4724cb49ffde8a: policy exists",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
