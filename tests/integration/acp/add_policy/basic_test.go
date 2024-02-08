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
				IsYAML: true,

				Creator: "cosmos1zzg43wdrhmmk89z3pmejwete2kkd4a3vn7w969",

				Policy: `
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

				ExpectedPolicyID: "dfe202ffb4f0fe9b46157c313213a3839e08a6f0a7c3aba55e4724cb49ffde8a",
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
				IsYAML: false,

				Creator: "cosmos1zzg43wdrhmmk89z3pmejwete2kkd4a3vn7w969",

				Policy: `
					{
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

				ExpectedPolicyID: "dfe202ffb4f0fe9b46157c313213a3839e08a6f0a7c3aba55e4724cb49ffde8a",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
