// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_dac_link_collection

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	schemaUtils "github.com/sourcenetwork/defradb/tests/integration/collection_version"
)

func TestACP_LinkCollection_WithMultipleResources_AcceptCollection(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{

			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a policy
name: test
resources:
- name: books
  permissions:
  - name: delete
  - name: read
  - name: update
- name: users
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - name: reader
    types:
    - actor
`,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
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
					"__type": map[string]any{
						"name": "Users", // NOTE: "Users" MUST exist
						"fields": schemaUtils.DefaultFields.Append(
							schemaUtils.Field{
								"name": "name",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "String",
								},
							},
						).Append(
							schemaUtils.Field{
								"name": "age",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "Int",
								},
							},
						).Tidy(),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_LinkCollection_WithMultipleResourcesBothBeingUsed_AcceptCollection(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{

			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a policy
name: test
resources:
- name: books
  permissions:
  - name: delete
  - name: read
  - name: update
- name: users
  permissions:
  - name: delete
  - expr: reader
    name: read
  - name: update
  relations:
  - name: reader
    types:
    - actor
`,
			},

			&action.AddCollection{
				SDL: `
					type Users @policy(
						id: "{{.Policy0}}",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
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
					"__type": map[string]any{
						"name": "Users", // NOTE: "Users" MUST exist
						"fields": schemaUtils.DefaultFields.Append(
							schemaUtils.Field{
								"name": "name",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "String",
								},
							},
						).Append(
							schemaUtils.Field{
								"name": "age",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "Int",
								},
							},
						).Tidy(),
					},
				},
			},

			&action.AddCollection{
				SDL: `
					type Books @policy(
						id: "{{.Policy0}}",,
						resource: "books"
					) {
						name: String
					}
				`,
			},

			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "Books") {
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
					"__type": map[string]any{
						"name": "Books", // NOTE: "Books" MUST exist
						"fields": schemaUtils.DefaultFields.Append(
							schemaUtils.Field{
								"name": "name",
								"type": map[string]any{
									"kind": "SCALAR",
									"name": "String",
								},
							},
						).Tidy(),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
