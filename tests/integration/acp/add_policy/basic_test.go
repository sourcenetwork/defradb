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

func TestACP_AddPolicy_BasicYAML_ValidPolicyID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, adding basic policy in YAML format",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
                    name: test
                    description: a basic policy that satisfies minimum DPI requirements

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

				ExpectedPolicyID: "aa664afaf8dff947ba85f4d464662d595af6c1e2466bd11fd6b82ea95b547ea3",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_AddPolicy_BasicJSON_ValidPolicyID(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, adding basic policy in JSON format",

		Actions: []any{
			testUtils.AddPolicy{
				Identity: actor1Identity,

				Policy: `
					{
					  "name": "test",
					  "description": "a basic policy that satisfies minimum DPI requirements",
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
					  },
					  "actor": {
					    "name": "actor"
					  }
					}
                `,

				ExpectedPolicyID: "aa664afaf8dff947ba85f4d464662d595af6c1e2466bd11fd6b82ea95b547ea3",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
