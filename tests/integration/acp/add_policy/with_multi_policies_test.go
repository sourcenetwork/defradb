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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_AddPolicy_AddMultipleDifferentPolicies_ValidPolicyIDs(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add multiple different policies",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: testUtils.ClientIdentity(1),

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
			},

			testUtils.AddPolicy{
				Identity: testUtils.ClientIdentity(1),

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
				Identity: testUtils.ClientIdentity(1),

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

				ExpectedPolicyID: immutable.Some(
					"468a5f345b3afec72f025185159e0fe84b02ead3374ec6aa54c390b2e3299444",
				),
			},

			testUtils.AddPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test2
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

				ExpectedPolicyID: immutable.Some(
					"47f986f872dfcfc7d126b017a4258df86d508f3333d93f19c0eb7f4624e97023",
				),
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
				Identity: testUtils.ClientIdentity(1),

				Policy: policyUsedByBoth,

				ExpectedPolicyID: immutable.Some(
					"468a5f345b3afec72f025185159e0fe84b02ead3374ec6aa54c390b2e3299444",
				),
			},

			testUtils.AddPolicy{
				Identity: testUtils.ClientIdentity(2),

				Policy: policyUsedByBoth,

				ExpectedPolicyID: immutable.Some(
					"66f3e364004a181e9b129f65dea317322d2285226e926d7e8cdfd644954e4262",
				),
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
				Identity: testUtils.ClientIdentity(1),

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

				ExpectedPolicyID: immutable.Some(
					"468a5f345b3afec72f025185159e0fe84b02ead3374ec6aa54c390b2e3299444",
				),
			},

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
                          read:
                            expr: owner
                          write:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor

                `,

				ExpectedPolicyID: immutable.Some(
					"66f3e364004a181e9b129f65dea317322d2285226e926d7e8cdfd644954e4262",
				),
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
				Identity: testUtils.ClientIdentity(1),

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

				ExpectedPolicyID: immutable.Some(
					"468a5f345b3afec72f025185159e0fe84b02ead3374ec6aa54c390b2e3299444",
				),
			},

			testUtils.AddPolicy{
				Identity: testUtils.ClientIdentity(1),

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

				ExpectedPolicyID: immutable.Some(
					"66f3e364004a181e9b129f65dea317322d2285226e926d7e8cdfd644954e4262",
				),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
