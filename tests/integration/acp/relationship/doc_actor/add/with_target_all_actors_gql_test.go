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

func TestACP_OwnerGivesOnlyReadAccessToAllActors_GQL_AllActorsCanReadButNotUpdateOrDelete(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, owner gives read access to all actors (gql), but the other actor can't update or delete",

		SupportedMutationTypes: immutable.Some(
			[]testUtils.MutationType{
				// GQL mutation will return no error when wrong identity is used (only for update requests),
				// so test that separately.
				testUtils.GQLRequestMutationType,
			},
		),

		Actions: []any{
			testUtils.AddPolicy{

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
				Identity: testUtils.ClientIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(2), // This identity can not read yet.

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

			testUtils.DeleteDoc{ // Since it can't read, it can't delete either.
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2),

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.UpdateDoc{ // Since it can't read, it can't update either.
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2),

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				SkipLocalUpdateEvent: true,
			},

			testUtils.AddDocActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.AllClientIdentities(),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(2), // Now any identity can read

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

			testUtils.Request{
				Identity: testUtils.ClientIdentity(3), // Now any identity can read

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

			testUtils.UpdateDoc{ // But doesn't mean they can update.
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2),

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.DeleteDoc{ // But doesn't mean they can delete.
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2),

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_OwnerGivesOnlyReadAccessToAllActors_GQL_CanReadEvenWithoutIdentityButNotUpdateOrDelete(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, owner gives read access to all actors (gql), can read without an identity but can't update or delete",

		SupportedMutationTypes: immutable.Some(
			[]testUtils.MutationType{
				// GQL mutation will return no error when wrong identity is used (only for update requests),
				// so test that separately.
				testUtils.GQLRequestMutationType,
			},
		),

		Actions: []any{
			testUtils.AddPolicy{

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
				Identity: testUtils.ClientIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.Request{
				Identity: testUtils.NoIdentity(), // Can not read without an identity.

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

			testUtils.DeleteDoc{ // Since can't read without identity, can't delete either.
				CollectionID: 0,

				Identity: testUtils.NoIdentity(),

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.UpdateDoc{ // Since can't read without identity, can't update either.
				CollectionID: 0,

				Identity: testUtils.NoIdentity(),

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				SkipLocalUpdateEvent: true,
			},

			testUtils.AddDocActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.AllClientIdentities(),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			testUtils.Request{
				Identity: testUtils.NoIdentity(), // Now any identity can read, even if there is no identity

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

			testUtils.UpdateDoc{ // But doesn't mean they can update.
				CollectionID: 0,

				Identity: testUtils.NoIdentity(),

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.DeleteDoc{ // But doesn't mean they can delete.
				CollectionID: 0,

				Identity: testUtils.NoIdentity(),

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
