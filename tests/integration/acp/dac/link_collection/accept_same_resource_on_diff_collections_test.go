// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package test_acp_dac_link_collection

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	schemaUtils "github.com/sourcenetwork/defradb/tests/integration/collection_version"
)

func TestACP_LinkCollection_UseSameResourceOnDifferentSchemas_AcceptCollections(t *testing.T) {
	sharedSameResourceName := "users"

	test := testUtils.TestCase{

		Actions: []any{

			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: a policy
name: test
resources:
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
				SDL: fmt.Sprintf(`
					type OldUsers @policy(
						id: "{{.Policy0}}",
						resource: "%s"
					) {
						name: String
						age: Int
					}
				`,
					sharedSameResourceName,
				),
			},

			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "OldUsers") {
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
						"name": "OldUsers", // NOTE: "OldUsers" MUST exist
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
				SDL: fmt.Sprintf(`
					type NewUsers @policy(
						id: "{{.Policy0}}",
						resource: "%s"
					) {
						name: String
						age: Int
					}
				`,
					sharedSameResourceName,
				),
			},

			testUtils.IntrospectionRequest{
				Request: `
					query {
						__type (name: "NewUsers") {
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
						"name": "NewUsers", // NOTE: "NewUsers" MUST exist
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
