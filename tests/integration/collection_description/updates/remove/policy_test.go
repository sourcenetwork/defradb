// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package remove

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestColDescrUpdateRemovePolicy_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.AddPolicy{

				Identity: immutable.Some(1),

				Policy: `
                    name: test
                    description: a test policy which marks a collection in a database as a resource

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader
                          write:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                          admin:
                            manages:
                              - reader
                            types:
                              - actor
                `,

				ExpectedPolicyID: "94eb195c0e459aa79e02a1986c7e731c5015721c18a373f2b2a0ed140a04b454",
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "94eb195c0e459aa79e02a1986c7e731c5015721c18a373f2b2a0ed140a04b454",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "remove", "path": "/1/Policy" }
					]
				`,
				ExpectedError: "collection policy cannot be mutated. CollectionID: 1",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
