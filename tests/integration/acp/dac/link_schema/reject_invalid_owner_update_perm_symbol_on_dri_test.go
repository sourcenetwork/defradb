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

func TestACP_LinkSchema_OwnerRelationWithDifferenceSetOpOnUpdatePermissionExprOnDRI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{

			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
actor:
  name: actor
description: a policy
name: test
resources:
- name: users
  permissions:
  - expr: owner
    name: delete
  - expr: owner + reader
    name: read
  - expr: owner - reader
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
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,

				ExpectedError: fmt.Sprintf(
					"expr of required permission has invalid character after relation. Permission: %s, Relation: %s, Character: %s",
					"update",
					"owner",
					"-",
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

func TestACP_LinkSchema_OwnerRelationWithIntersectionSetOpOnUpdatePermissionExprOnDRI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{

			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
actor:
  name: actor
description: a policy
name: test
resources:
- name: users
  permissions:
  - expr: owner
    name: delete
  - expr: owner + reader
    name: read
  - expr: owner & reader
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
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,

				ExpectedError: fmt.Sprintf(
					"expr of required permission has invalid character after relation. Permission: %s, Relation: %s, Character: %s",
					"update",
					"owner",
					"&",
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

func TestACP_LinkSchema_OwnerRelationWithInvalidSetOpOnUpdatePermissionExprOnDRI_SchemaRejected(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{

			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
actor:
  name: actor
description: a policy
name: test
resources:
- name: users
  permissions:
  - expr: owner
    name: delete
  - expr: owner
    name: read
  - expr: owner - owner
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
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,

				ExpectedError: fmt.Sprintf(
					"expr of required permission has invalid character after relation. Permission: %s, Relation: %s, Character: %s",
					"update",
					"owner",
					"-",
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
