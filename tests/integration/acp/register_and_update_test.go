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

func TestACP_CreateWithoutIdentityAndUpdateWithoutIdentity_CanUpdate(t *testing.T) {
	// The same identity that is used to do the registering/creation should be used in the
	// final read check to see the state of that registered document.
	// Note: In this test that identity is empty (no identity).

	test := testUtils.TestCase{

		Description: "Test acp, create without identity, and update without identity, can update",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: Actor1Identity,

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
				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
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
						"_docID": "bae-1e608f7d-b01e-5dd5-ad4a-9c6cc3005a36",
						"name":   "Shahzad Lone",
						"age":    int64(28),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CreateWithoutIdentityAndUpdateWithIdentity_CanUpdate(t *testing.T) {
	// The same identity that is used to do the registering/creation should be used in the
	// final read check to see the state of that registered document.
	// Note: In this test that identity is empty (no identity).

	test := testUtils.TestCase{

		Description: "Test acp, create without identity, and update with identity, can update",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: Actor1Identity,

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
				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: Actor1Identity,

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
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
						"_docID": "bae-1e608f7d-b01e-5dd5-ad4a-9c6cc3005a36",
						"name":   "Shahzad Lone",
						"age":    int64(28),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CreateWithIdentityAndUpdateWithIdentity_CanUpdate(t *testing.T) {
	// OwnerIdentity should be the same identity that is used to do the registering/creation,
	// and the final read check to see the state of that registered document.
	OwnerIdentity := Actor1Identity

	test := testUtils.TestCase{

		Description: "Test acp, create with identity, and update with identity, can update",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: OwnerIdentity,

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
				CollectionID: 0,

				Identity: OwnerIdentity,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: OwnerIdentity,

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
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
						"name":   "Shahzad Lone",
						"age":    int64(28),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CreateWithIdentityAndUpdateWithoutIdentity_CanNotUpdate(t *testing.T) {
	// OwnerIdentity should be the same identity that is used to do the registering/creation,
	// and the final read check to see the state of that registered document.
	OwnerIdentity := Actor1Identity

	test := testUtils.TestCase{

		Description: "Test acp, create with identity, and update without identity, can not update",

		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL mutation will return no error when wrong identity is used so test that separately.
			testUtils.CollectionNamedMutationType,
			testUtils.CollectionSaveMutationType,
		}),

		Actions: []any{
			testUtils.AddPolicy{

				Identity: OwnerIdentity,

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
				CollectionID: 0,

				Identity: OwnerIdentity,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

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

func TestACP_CreateWithIdentityAndUpdateWithWrongIdentity_CanNotUpdate(t *testing.T) {
	// OwnerIdentity should be the same identity that is used to do the registering/creation,
	// and the final read check to see the state of that registered document.
	OwnerIdentity := Actor1Identity

	WrongIdentity := Actor2Identity

	test := testUtils.TestCase{

		Description: "Test acp, create with identity, and update without identity, can not update",

		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL mutation will return no error when wrong identity is used so test that separately.
			testUtils.CollectionNamedMutationType,
			testUtils.CollectionSaveMutationType,
		}),

		Actions: []any{
			testUtils.AddPolicy{

				Identity: OwnerIdentity,

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
				CollectionID: 0,

				Identity: OwnerIdentity,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: WrongIdentity,

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

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

// This separate GQL test should be merged with the ones above when all the clients are fixed
// to behave the same in: https://github.com/sourcenetwork/defradb/issues/2410
func TestACP_CreateWithIdentityAndUpdateWithoutIdentityGQL_CanNotUpdate(t *testing.T) {
	// OwnerIdentity should be the same identity that is used to do the registering/creation,
	// and the final read check to see the state of that registered document.
	OwnerIdentity := Actor1Identity

	test := testUtils.TestCase{

		Description: "Test acp, create with identity, and update without identity (gql), can not update",

		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL mutation will return no error when wrong identity is used so test that separately.
			testUtils.GQLRequestMutationType,
		}),

		Actions: []any{
			testUtils.AddPolicy{

				Identity: OwnerIdentity,

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
				CollectionID: 0,

				Identity: OwnerIdentity,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
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

// This separate GQL test should be merged with the ones above when all the clients are fixed
// to behave the same in: https://github.com/sourcenetwork/defradb/issues/2410
func TestACP_CreateWithIdentityAndUpdateWithWrongIdentityGQL_CanNotUpdate(t *testing.T) {
	// OwnerIdentity should be the same identity that is used to do the registering/creation,
	// and the final read check to see the state of that registered document.
	OwnerIdentity := Actor1Identity

	WrongIdentity := Actor2Identity

	test := testUtils.TestCase{

		Description: "Test acp, create with identity, and update without identity (gql), can not update",

		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL mutation will return no error when wrong identity is used so test that separately.
			testUtils.GQLRequestMutationType,
		}),

		Actions: []any{
			testUtils.AddPolicy{

				Identity: OwnerIdentity,

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
				CollectionID: 0,

				Identity: OwnerIdentity,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: WrongIdentity,

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
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
