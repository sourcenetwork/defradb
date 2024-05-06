// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_index

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	acpUtils "github.com/sourcenetwork/defradb/tests/integration/acp"
)

// This test documents that we don't allow creating indexes on collections that have policy
// until the following is implemented:
// TODO-ACP: ACP <> P2P https://github.com/sourcenetwork/defradb/issues/2365
func TestACP_IndexCreateWithSeparateRequest_OnCollectionWithPolicy_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, with creating new index using separate request on permissioned collection, error",
		Actions: []any{

			testUtils.AddPolicy{

				Identity: acpUtils.Actor1Identity,

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

			testUtils.CreateIndex{
				CollectionID: 0,

				IndexName: "some_index",

				FieldName: "name",

				ExpectedError: "can not create index on a collection with a policy",
			},

			testUtils.Request{
				Request: `
					query  {
						Users {
							name
							age
						}
					}`,

				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test documents that we don't allow creating indexes on collections that have policy
// until the following is implemented:
// TODO-ACP: ACP <> P2P https://github.com/sourcenetwork/defradb/issues/2365
func TestACP_IndexCreateWithDirective_OnCollectionWithPolicy_ReturnError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test acp, with creating new index using directive on permissioned collection, error",
		Actions: []any{

			testUtils.AddPolicy{

				Identity: acpUtils.Actor1Identity,

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
						name: String @index
						age: Int
					}
				`,

				ExpectedError: "can not create index on a collection with a policy",
			},

			testUtils.Request{
				Request: `
					query  {
						Users {
							name
							age
						}
					}`,

				ExpectedError: `Cannot query field "Users" on type "Query"`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
