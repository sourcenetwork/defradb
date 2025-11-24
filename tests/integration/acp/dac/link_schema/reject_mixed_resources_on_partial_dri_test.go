// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_dac_link_schema

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_LinkSchema_UseInvalidResource_RejectSchema(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{

			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
actor:
  name: actor
description: A Partially DRI Compliant Policy
name: test
resources:
- name: usersInvalid
  permissions:
  - expr: reader
    name: delete
  - expr: reader - owner
    name: read
  - expr: reader
    name: update
  relations:
  - name: owner
    types:
    - actor
  - name: reader
    types:
    - actor
- name: usersValid
  permissions:
  - expr: owner
    name: delete
  - expr: owner + reader
    name: read
  - expr: owner
    name: update
  relations:
  - name: owner
    types:
    - actor
  - name: reader
    types:
    - actor
`,
			},

			&action.AddSchema{
				Schema: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "usersInvalid"
					) {
						name: String
						age: Int
					}
				`,

				ExpectedError: fmt.Sprintf(
					"expr of required permission must start with required relation. Permission: %s, Relation: %s",
					"read",
					"owner",
				),
			},

			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "Users") {
							name
							fields {
								name
								type {
								name
								kind
								}
							}
						}
					}
				`,
				ExpectedData: map[string]any{
					"__type": nil, // NOTE: No "Users" should exist.
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
