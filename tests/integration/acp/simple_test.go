// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_CreateAndRead(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Simple acp create and read",

		Actions: []any{
			testUtils.AddPolicy{
				IsYAML: true,

				Creator: "cosmos1zzg43wdrhmmk89z3pmejwete2kkd4a3vn7w969",

				Policy: `
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

				ExpectedPolicyID: "53980e762616fcffbe76307995895e862f87ef3f21d509325d1dc772a770b001",
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "53980e762616fcffbe76307995895e862f87ef3f21d509325d1dc772a770b001",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": 27
				}`,
			},

			testUtils.Request{
				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,
				Results: []map[string]any{
					{
						"_docID": "bae-88b63198-7d38-5714-a9ff-21ba46374fd1",
						"name":   "John",
						"age":    int64(27),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
