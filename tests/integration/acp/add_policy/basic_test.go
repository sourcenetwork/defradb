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
				Identity: testUtils.ClientIdentity(1),

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

				ExpectedPolicyID: "468a5f345b3afec72f025185159e0fe84b02ead3374ec6aa54c390b2e3299444",
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
				Identity: testUtils.ClientIdentity(1),

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

				ExpectedPolicyID: "468a5f345b3afec72f025185159e0fe84b02ead3374ec6aa54c390b2e3299444",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
