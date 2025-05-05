// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_dac_relationship_doc_actor_delete

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_OwnerRevokesUpdateAccess_OtherActorCanNoLongerUpdate(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, owner revokes update access from another actor, they can not update anymore",

		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			testUtils.CollectionNamedMutationType,
			testUtils.CollectionSaveMutationType,
		}),

		Actions: []any{
			testUtils.AddDocPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: Test Policy

                    description: A Policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader + updater + deleter

                          update:
                            expr: owner + updater

                          delete:
                            expr: owner + deleter

                          nothing:
                            expr: dummy

                        relations:
                          owner:
                            types:
                              - actor

                          reader:
                            types:
                              - actor

                          updater:
                            types:
                              - actor

                          deleter:
                            types:
                              - actor

                          admin:
                            manages:
                              - reader
                            types:
                              - actor

                          dummy:
                            types:
                              - actor
                `,
			},

			testUtils.SchemaUpdate{
				Schema: `
						type Users @policy(
							id: "{{.Policy0}}",
							resource: "users"
						) {
							name: String
							age: Int
						}
					`,

				Replace: map[string]testUtils.ReplaceType{
					"Policy0": testUtils.NewPolicyIndex(0),
				},
			},

			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			// Give access to the other actor to update and read the document.
			testUtils.AddDocActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "updater",

				ExpectedExistence: false,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2), // This identity can update.

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
			},

			// Ensure the other identity can read and update the document.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(2), // This identity can also read.

				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,

				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad Lone",
							"age":  int64(28),
						},
					},
				},
			},

			testUtils.DeleteDocActorRelationship{ // Revoke access from being able to update (and read) the document.
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "updater",

				ExpectedRecordFound: true,
			},

			// The other identity can neither update nor read the other document anymore.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(2),

				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,

				Results: map[string]any{
					"Users": []map[string]any{}, // Can't read the document anymore
				},
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2),

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Update Not Possible"
					}
				`,

				ExpectedError: "document not found or not authorized to access", // Can't update the document anymore.
			},

			// Ensure document was not accidentally updated using owner identity.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(1),

				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,

				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad Lone",
							"age":  int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_OwnerRevokesUpdateAccess_GQL_OtherActorCanNoLongerUpdate(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, owner revokes update access from another actor, they can not update anymore (gql)",

		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL mutation will return no error.
			testUtils.GQLRequestMutationType,
		}),

		Actions: []any{
			testUtils.AddDocPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: Test Policy

                    description: A Policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader + updater + deleter

                          update:
                            expr: owner + updater

                          delete:
                            expr: owner + deleter

                          nothing:
                            expr: dummy

                        relations:
                          owner:
                            types:
                              - actor

                          reader:
                            types:
                              - actor

                          updater:
                            types:
                              - actor

                          deleter:
                            types:
                              - actor

                          admin:
                            manages:
                              - reader
                            types:
                              - actor

                          dummy:
                            types:
                              - actor
                `,
			},

			testUtils.SchemaUpdate{
				Schema: `
						type Users @policy(
							id: "{{.Policy0}}",
							resource: "users"
						) {
							name: String
							age: Int
						}
					`,

				Replace: map[string]testUtils.ReplaceType{
					"Policy0": testUtils.NewPolicyIndex(0),
				},
			},

			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			// Give access to the other actor to update and read the document.
			testUtils.AddDocActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "updater",

				ExpectedExistence: false,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2), // This identity can update.

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
			},

			// Ensure the other identity can read and update the document.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(2), // This identity can also read.

				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,

				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad Lone",
							"age":  int64(28),
						},
					},
				},
			},

			testUtils.DeleteDocActorRelationship{ // Revoke access from being able to update (and read) the document.
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "updater",

				ExpectedRecordFound: true,
			},

			// The other identity can neither update nor read the other document anymore.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(2),

				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,

				Results: map[string]any{
					"Users": []map[string]any{}, // Can't read the document anymore
				},
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2),

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Update Not Possible"
					}
				`,

				SkipLocalUpdateEvent: true,
			},

			// Ensure document was not accidentally updated using owner identity.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(1),

				Request: `
					query {
						Users {
							name
							age
						}
					}
				`,

				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Shahzad Lone",
							"age":  int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
