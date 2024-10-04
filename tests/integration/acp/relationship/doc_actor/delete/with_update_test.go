// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_relationship_doc_actor_delete

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_OwnerRevokesUpdateWriteAccess_OtherActorCanNoLongerUpdate(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, owner revokes write(update) access from another actor, they can not update anymore",

		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			testUtils.CollectionNamedMutationType,
			testUtils.CollectionSaveMutationType,
		}),

		Actions: []any{
			testUtils.AddPolicy{

				Identity: immutable.Some(1),

				Policy: `
                    name: Test Policy

                    description: A Policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader + writer

                          write:
                            expr: owner + writer

                          nothing:
                            expr: dummy

                        relations:
                          owner:
                            types:
                              - actor

                          reader:
                            types:
                              - actor

                          writer:
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

				ExpectedPolicyID: expectedPolicyID,
			},

			testUtils.SchemaUpdate{
				Schema: fmt.Sprintf(`
						type Users @policy(
							id: "%s",
							resource: "users"
						) {
							name: String
							age: Int
						}
					`,
					expectedPolicyID,
				),
			},

			testUtils.CreateDoc{
				Identity: immutable.Some(1),

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
				RequestorIdentity: 1,

				TargetIdentity: 2,

				CollectionID: 0,

				DocID: 0,

				Relation: "writer",

				ExpectedExistence: false,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: immutable.Some(2), // This identity can update.

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
			},

			// Ensure the other identity can read and update the document.
			testUtils.Request{
				Identity: immutable.Some(2), // This identity can also read.

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
				RequestorIdentity: 1,

				TargetIdentity: 2,

				CollectionID: 0,

				DocID: 0,

				Relation: "writer",

				ExpectedRecordFound: true,
			},

			// The other identity can neither update nor read the other document anymore.
			testUtils.Request{
				Identity: immutable.Some(2),

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

				Identity: immutable.Some(2),

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
				Identity: immutable.Some(1),

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

func TestACP_OwnerRevokesUpdateWriteAccess_GQL_OtherActorCanNoLongerUpdate(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, owner revokes write(update) access from another actor, they can not update anymore (gql)",

		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL mutation will return no error.
			testUtils.GQLRequestMutationType,
		}),

		Actions: []any{
			testUtils.AddPolicy{

				Identity: immutable.Some(1),

				Policy: `
                    name: Test Policy

                    description: A Policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader + writer

                          write:
                            expr: owner + writer

                          nothing:
                            expr: dummy

                        relations:
                          owner:
                            types:
                              - actor

                          reader:
                            types:
                              - actor

                          writer:
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

				ExpectedPolicyID: expectedPolicyID,
			},

			testUtils.SchemaUpdate{
				Schema: fmt.Sprintf(`
						type Users @policy(
							id: "%s",
							resource: "users"
						) {
							name: String
							age: Int
						}
					`,
					expectedPolicyID,
				),
			},

			testUtils.CreateDoc{
				Identity: immutable.Some(1),

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
				RequestorIdentity: 1,

				TargetIdentity: 2,

				CollectionID: 0,

				DocID: 0,

				Relation: "writer",

				ExpectedExistence: false,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: immutable.Some(2), // This identity can update.

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
			},

			// Ensure the other identity can read and update the document.
			testUtils.Request{
				Identity: immutable.Some(2), // This identity can also read.

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
				RequestorIdentity: 1,

				TargetIdentity: 2,

				CollectionID: 0,

				DocID: 0,

				Relation: "writer",

				ExpectedRecordFound: true,
			},

			// The other identity can neither update nor read the other document anymore.
			testUtils.Request{
				Identity: immutable.Some(2),

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

				Identity: immutable.Some(2),

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
				Identity: immutable.Some(1),

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
