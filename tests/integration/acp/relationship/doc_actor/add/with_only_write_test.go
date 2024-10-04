// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_relationship_doc_actor_add

import (
	"fmt"
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_OwnerGivesUpdateWriteAccessToAnotherActorWithoutExplicitReadPerm_OtherActorCantUpdate(t *testing.T) {
	expectedPolicyID := "0a243b1e61f990bccde41db7e81a915ffa1507c1403ae19727ce764d3b08846b"

	test := testUtils.TestCase{

		Description: "Test acp, owner gives write(update) access to another actor, without explicit read permission",

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
                            expr: owner + reader

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

			testUtils.Request{
				Identity: immutable.Some(2), // This identity can not read yet.

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
					"Users": []map[string]any{}, // Can't see the documents yet
				},
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: immutable.Some(2), // This identity can not update yet.

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				ExpectedError: "document not found or not authorized to access",
			},

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

				Identity: immutable.Some(2), // This identity can still not update.

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.Request{
				Identity: immutable.Some(2), // This identity can still not read.

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

func TestACP_OwnerGivesDeleteWriteAccessToAnotherActorWithoutExplicitReadPerm_OtherActorCantDelete(t *testing.T) {
	expectedPolicyID := "0a243b1e61f990bccde41db7e81a915ffa1507c1403ae19727ce764d3b08846b"

	test := testUtils.TestCase{

		Description: "Test acp, owner gives write(delete) access to another actor, without explicit read permission",

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
                            expr: owner + reader

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

			testUtils.Request{
				Identity: immutable.Some(2), // This identity can not read yet.

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
					"Users": []map[string]any{}, // Can't see the documents yet
				},
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: immutable.Some(2), // This identity can not delete yet.

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.AddDocActorRelationship{
				RequestorIdentity: 1,

				TargetIdentity: 2,

				CollectionID: 0,

				DocID: 0,

				Relation: "writer",

				ExpectedExistence: false,
			},

			testUtils.Request{
				Identity: immutable.Some(2), // This identity can still not read.

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

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: immutable.Some(2), // This identity can still not delete.

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
