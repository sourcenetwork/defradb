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

func TestACP_CreateWithoutIdentityAndDeleteWithoutIdentity_CanDelete(t *testing.T) {
	// The same identity that is used to do the registering/creation should be used in the
	// final read check to see the state of that registered document.
	// Note: In this test that identity is empty (no identity).

	test := testUtils.TestCase{

		Description: "Test acp, create without identity, and delete without identity, can delete",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: Actor1Identity,

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

				ExpectedPolicyID: "7bcb558ef8dac6b744a11ea144a61a756ea38475554097ac04612037c36ffe52",
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "7bcb558ef8dac6b744a11ea144a61a756ea38475554097ac04612037c36ffe52",
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

				Results: []map[string]any{},
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

				Identity: Actor1Identity,

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

				ExpectedPolicyID: "7bcb558ef8dac6b744a11ea144a61a756ea38475554097ac04612037c36ffe52",
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "7bcb558ef8dac6b744a11ea144a61a756ea38475554097ac04612037c36ffe52",
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

				Identity: Actor1Identity,

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
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CreateWithIdentityAndDeleteWithIdentity_CanDelete(t *testing.T) {
	// OwnerIdentity should be the same identity that is used to do the registering/creation,
	// and the final read check to see the state of that registered document.
	OwnerIdentity := Actor1Identity

	test := testUtils.TestCase{

		Description: "Test acp, create with identity, and delete with identity, can delete",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: OwnerIdentity,

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

				ExpectedPolicyID: "7bcb558ef8dac6b744a11ea144a61a756ea38475554097ac04612037c36ffe52",
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "7bcb558ef8dac6b744a11ea144a61a756ea38475554097ac04612037c36ffe52",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			testUtils.CreateDoc{
				CollectionID: 0,

				Identity: OwnerIdentity,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: OwnerIdentity,

				DocID: 0,
			},

			testUtils.Request{
				Identity: OwnerIdentity,

				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,
				Results: []map[string]any{},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CreateWithIdentityAndDeleteWithoutIdentity_CanNotDelete(t *testing.T) {
	// OwnerIdentity should be the same identity that is used to do the registering/creation,
	// and the final read check to see the state of that registered document.
	OwnerIdentity := Actor1Identity

	test := testUtils.TestCase{

		Description: "Test acp, create with identity, and delete without identity, can not delete",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: OwnerIdentity,

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

				ExpectedPolicyID: "7bcb558ef8dac6b744a11ea144a61a756ea38475554097ac04612037c36ffe52",
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "7bcb558ef8dac6b744a11ea144a61a756ea38475554097ac04612037c36ffe52",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			testUtils.CreateDoc{
				CollectionID: 0,

				Identity: OwnerIdentity,

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
				Identity: OwnerIdentity,

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
						"_docID": "bae-1e608f7d-b01e-5dd5-ad4a-9c6cc3005a36",
						"name":   "Shahzad",
						"age":    int64(28),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CreateWithIdentityAndDeleteWithWrongIdentity_CanNotDelete(t *testing.T) {
	// OwnerIdentity should be the same identity that is used to do the registering/creation,
	// and the final read check to see the state of that registered document.
	OwnerIdentity := Actor1Identity

	WrongIdentity := Actor2Identity

	test := testUtils.TestCase{

		Description: "Test acp, create with identity, and delete without identity, can not delete",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: OwnerIdentity,

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

				ExpectedPolicyID: "7bcb558ef8dac6b744a11ea144a61a756ea38475554097ac04612037c36ffe52",
			},

			testUtils.SchemaUpdate{
				Schema: `
					type Users @policy(
						id: "7bcb558ef8dac6b744a11ea144a61a756ea38475554097ac04612037c36ffe52",
						resource: "users"
					) {
						name: String
						age: Int
					}
				`,
			},

			testUtils.CreateDoc{
				CollectionID: 0,

				Identity: OwnerIdentity,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: WrongIdentity,

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.Request{
				Identity: OwnerIdentity,

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
						"_docID": "bae-1e608f7d-b01e-5dd5-ad4a-9c6cc3005a36",
						"name":   "Shahzad",
						"age":    int64(28),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
