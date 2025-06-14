// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_dac_add_policy

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_AddPolicy_MultipleResources_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, multiple resources, valid ID",

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          update:
                            expr: owner
                          delete:
                            expr: owner
                          read:
                            expr: owner + reader

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                      books:
                        permissions:
                          update:
                            expr: owner
                          delete:
                            expr: owner
                          read:
                            expr: owner + reader

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                `,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_MultipleResourcesUsingRelationDefinedInOther_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, multiple resources using other's relation, return error",

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          update:
                            expr: owner
                          delete:
                            expr: owner
                          read:
                            expr: owner + reader

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                      books:
                        permissions:
                          update:
                            expr: owner
                          delete:
                            expr: owner
                          read:
                            expr: owner + reader

                        relations:
                          owner:
                            types:
                              - actor
                `,

				ExpectedError: "resource books missing relation reader",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_SecondResourcesMissingRequiredOwner_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, multiple resources second missing required owner, return error",

		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
                    description: a policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          update:
                            expr: owner
                          delete:
                            expr: owner
                          read:
                            expr: owner + reader

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                      books:
                        permissions:
                          update:
                            expr: owner
                          delete:
                            expr: owner
                          read:
                            expr: owner + reader

                        relations:
                          reader:
                            types:
                              - actor
                `,

				ExpectedError: "BAD_INPUT",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
