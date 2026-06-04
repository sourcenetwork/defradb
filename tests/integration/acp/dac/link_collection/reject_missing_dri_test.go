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
)

func TestACP_LinkCollection_WhereNoPolicyWasAdded_CollectionRejected(t *testing.T) {
	nonExistingPolicyID := "1239a04400966b311339f62db50044b1bde70cece2ce9897d69c1bafa5cfab81"

	test := testUtils.TestCase{

		Actions: []any{

			&action.AddCollection{
				SDL: fmt.Sprintf(`
					type Users @policy(
						id: "%s",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
					nonExistingPolicyID,
				),

				ExpectedError: "policyID specified does not exist with acp",
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

func TestACP_LinkCollection_WhereAPolicyWasAddedButLinkedPolicyWasNotAdded_CollectionRejected(t *testing.T) {
	incorrectPolicyID := "1239a04400966b311339f62db50044b1bde70cece2ce9897d69c1bafa5cfab81"

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
					type Users @policy(
						id: "%s",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
					incorrectPolicyID,
				),

				ExpectedError: "policyID specified does not exist with acp",
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
