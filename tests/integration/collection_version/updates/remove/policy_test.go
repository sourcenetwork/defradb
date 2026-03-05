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

package remove

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColVersionUpdateRemovePolicy_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: test
                    description: a test policy which marks a collection in a database as a resource

                    resources:
                    - name: users
                      permissions:
                      - name: read
                        expr: reader
                      - name: update
                      - name: delete
                      relations:
                      - name: reader
                        types:
                        - actor
                      - name: admin
                        manages:
                        - reader
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

			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "remove",
							"path": "/bafyreihsneodeja4lfer5puptim3lkwvketyckrmkhfpgxm67ch5wenjwq/Policy"
						}
					]
				`,
				ExpectedError: "collection policy cannot be mutated.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
