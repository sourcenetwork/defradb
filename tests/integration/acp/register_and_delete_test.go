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

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_CreateWithoutIdentityAndDeleteWithoutIdentity_CanDelete(t *testing.T) {
	// The same identity that is used to do the registering/creation should be used in the
	// final read check to see the state of that registered document.
	// Note: In this test that identity is empty (no identity).

	test := testUtils.TestCase{

		Description: "Test acp, create without identity, and delete without identity, can delete",

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

			testUtils.CreateDoc{
				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				DocID: 0,
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

				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CreateWithoutIdentityAndDeleteWithIdentity_CanDelete(t *testing.T) {
	// The same identity that is used to do the registering/creation should be used in the
	// final read check to see the state of that registered document.
	// Note: In this test that identity is empty (no identity).

	test := testUtils.TestCase{

		Description: "Test acp, create without identity, and delete with identity, can delete",

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

			testUtils.CreateDoc{
				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: immutable.Some(1),

				DocID: 0,
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
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CreateWithIdentityAndDeleteWithIdentity_CanDelete(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, create with identity, and delete with identity, can delete",

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

			testUtils.CreateDoc{
				CollectionID: 0,

				Identity: immutable.Some(1),

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: immutable.Some(1),

				DocID: 0,
			},

			testUtils.Request{
				Identity: immutable.Some(1),

				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CreateWithIdentityAndDeleteWithoutIdentity_CanNotDelete(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, create with identity, and delete without identity, can not delete",

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

			testUtils.CreateDoc{
				CollectionID: 0,

				Identity: immutable.Some(1),

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.Request{
				Identity: immutable.Some(1),

				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "bae-9d443d0c-52f6-568b-8f74-e8ff0825697b",
							"name":   "Shahzad",
							"age":    int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CreateWithIdentityAndDeleteWithWrongIdentity_CanNotDelete(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, create with identity, and delete without identity, can not delete",

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

			testUtils.CreateDoc{
				CollectionID: 0,

				Identity: immutable.Some(1),

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: immutable.Some(2),

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.Request{
				Identity: immutable.Some(1),

				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": "bae-9d443d0c-52f6-568b-8f74-e8ff0825697b",
							"name":   "Shahzad",
							"age":    int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
