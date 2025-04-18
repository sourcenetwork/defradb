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
                          update:
                            expr: owner
                          delete:
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
                          update:
                            expr: owner
                          delete:
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
                            "update": {
                              "expr": "owner"
                            },
                            "delete": {
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
					"60079fa5b415dfc6f6e6b70e123a8acb8de26d94d7ff9410449fb12950963ff0",
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
                          update:
                            expr: owner
                          delete:
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
					"32371d1285f8662ba54c8d63439f823b72e3347d517aa00cd7e305d73df57dcc",
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
			  update:
				expr: owner
			  delete:
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
					"60079fa5b415dfc6f6e6b70e123a8acb8de26d94d7ff9410449fb12950963ff0",
				),
			},

			testUtils.AddPolicy{
				Identity: testUtils.ClientIdentity(2),

				Policy: policyUsedByBoth,

				ExpectedPolicyID: immutable.Some(
					"4f113ea28e09992fdf6f3a8ccac8be8d8d39c932f48f54c42fff9c3513cd9a7a",
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
                          update:
                            expr: owner
                          delete:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor

                `,

				ExpectedPolicyID: immutable.Some(
					"60079fa5b415dfc6f6e6b70e123a8acb8de26d94d7ff9410449fb12950963ff0",
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
                          update:
                            expr: owner
                          delete:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor

                `,

				ExpectedPolicyID: immutable.Some(
					"4f113ea28e09992fdf6f3a8ccac8be8d8d39c932f48f54c42fff9c3513cd9a7a",
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
                          update:
                            expr: owner
                          delete:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor
                `,

				ExpectedPolicyID: immutable.Some(
					"60079fa5b415dfc6f6e6b70e123a8acb8de26d94d7ff9410449fb12950963ff0",
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
                            "update": {
                              "expr": "owner"
                            },
                            "delete": {
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
					"4f113ea28e09992fdf6f3a8ccac8be8d8d39c932f48f54c42fff9c3513cd9a7a",
				),
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
