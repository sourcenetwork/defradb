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

func TestACP_AddPolicy_MultipleResources_ValidID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, multiple resources, valid ID",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: immutable.Some(1),

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

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                      books:
                        permissions:
                          write:
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

				ExpectedPolicyID: "a9e1a113ccc2609d7f99a42531017f0fbc9b736640ec8ffc7f09a1e29583ca45",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_MultipleResourcesUsingRelationDefinedInOther_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, add policy, multiple resources using other's relation, return error",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: immutable.Some(1),

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

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                      books:
                        permissions:
                          write:
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
			testUtils.AddPolicy{
				Identity: immutable.Some(1),

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

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                      books:
                        permissions:
                          write:
                            expr: owner
                          read:
                            expr: owner + reader

                        relations:
                          reader:
                            types:
                              - actor
                `,

				ExpectedError: "resource books: resource missing owner relation: invalid policy",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
