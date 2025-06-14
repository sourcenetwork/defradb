// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_dac_relationship_doc_actor_add

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_ManagerGivesReadAccessToAnotherActor_OtherActorCanRead(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, owner gives read access to another actor",

		Actions: []any{
			testUtils.AddDACPolicy{

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

			testUtils.Request{
				Identity: testUtils.ClientIdentity(3), // This identity can not read yet.

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

			testUtils.AddDACActorRelationship{ // Make admin / manager
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedExistence: false,
			},

			testUtils.AddDACActorRelationship{ // Admin makes another actor a reader
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(3),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(3), // The other actor can read

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

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(3), // The other actor can not update

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(3), // The other actor can not delete

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_ManagerGivesWriteAccessToAnotherActor_OtherActorCanWrite(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, owner gives write (update and delete) access to another actor",

		Actions: []any{
			testUtils.AddDACPolicy{

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
                              - updater
                              - deleter
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

			testUtils.Request{
				Identity: testUtils.ClientIdentity(3), // This identity can not read yet.

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

			testUtils.AddDACActorRelationship{ // Make admin / manager
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedExistence: false,
			},

			testUtils.AddDACActorRelationship{ // Admin makes another actor an updater
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(3),

				CollectionID: 0,

				DocID: 0,

				Relation: "updater",

				ExpectedExistence: false,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(3), // The other actor can update

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(3), // The other actor can read

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
							"name":   "Shahzad Lone", // Updated name
							"age":    int64(28),
						},
					},
				},
			},

			testUtils.AddDACActorRelationship{ // Admin makes another actor a deleter
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(3),

				CollectionID: 0,

				DocID: 0,

				Relation: "deleter",

				ExpectedExistence: false,
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(3), // The other actor can delete

				DocID: 0,
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(3),

				Request: `
					query {
						Users {
							_docID
							name
							age
						}
					}
				`,

				Results: map[string]any{ // Check actually deleted
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_OwnerMakesAManagerThatGivesItSelfReadAccess_ManagerCanRead(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, owner makes a manager that gives itself read access",

		Actions: []any{
			testUtils.AddDACPolicy{

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

			testUtils.Request{
				Identity: testUtils.ClientIdentity(2), // This identity (to be manager) can not read yet.

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

			testUtils.AddDACActorRelationship{ // Make admin / manager
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedExistence: false,
			},

			testUtils.AddDACActorRelationship{ // Manager makes itself a reader
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(2), // Manager can read now

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

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2), // Manager still can't update

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2), // Manager still can't delete

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_OwnerMakesAManagerThatGivesItSelfReadAndWriteAccess_ManagerCanReadAndWrite(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, owner makes a manager that gives itself read and write (update and delete) access",

		SupportedMutationTypes: immutable.Some(
			[]testUtils.MutationType{
				// GQL mutation will return no error when wrong identity is used with gql (only for update requests),
				testUtils.CollectionNamedMutationType,
				testUtils.CollectionSaveMutationType,
			}),

		Actions: []any{
			testUtils.AddDACPolicy{

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
                              - updater
                              - deleter
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

			testUtils.Request{
				Identity: testUtils.ClientIdentity(2), // This identity (to be manager) can not read yet.

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

				Identity: testUtils.ClientIdentity(2), // Manager can't update yet.

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2), // Manager can't delete yet.

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.AddDACActorRelationship{ // Make admin / manager
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedExistence: false,
			},

			testUtils.AddDACActorRelationship{ // Manager makes itself an updater.
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "updater",

				ExpectedExistence: false,
			},

			// Note: It is not neccesary to make itself a reader, as becoming an updater allows reading.
			testUtils.AddDACActorRelationship{ // Manager makes itself a reader
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2), // Manager can now update.

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(2), // Manager can read now

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
							"name":   "Shahzad Lone",
							"age":    int64(28),
						},
					},
				},
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2), // Manager can still not delete yet.

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.AddDACActorRelationship{ // Manager makes itself a deleter.
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "deleter",

				ExpectedExistence: false,
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2), // Manager can now delete.

				DocID: 0,
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(2), // Make sure manager was able to delete the document.

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

func TestACP_ManagerAddsRelationshipWithRelationItDoesNotManageAccordingToPolicy_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, manager adds relationship with relation it does not manage according to policy, error",

		SupportedMutationTypes: immutable.Some(
			[]testUtils.MutationType{
				// GQL mutation will return no error when wrong identity is used with gql (only for update requests),
				testUtils.CollectionNamedMutationType,
				testUtils.CollectionSaveMutationType,
			}),

		Actions: []any{
			testUtils.AddDACPolicy{

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

			testUtils.AddDACActorRelationship{ // Make admin / manager
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedExistence: false,
			},

			testUtils.AddDACActorRelationship{ // Admin tries to make another actor an updater.
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(3),

				CollectionID: 0,

				DocID: 0,

				Relation: "updater",

				ExpectedError: "UNAUTHORIZED",
			},

			testUtils.AddDACActorRelationship{ // Admin tries to make another actor a deleter.
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(3),

				CollectionID: 0,

				DocID: 0,

				Relation: "deleter",

				ExpectedError: "UNAUTHORIZED",
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(3), // The other actor can't read

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

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(3), // The other actor can not update

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(3), // The other actor can not delete

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_OwnerMakesManagerButManagerCanNotPerformOperations_ManagerCantReadOrWrite(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, owner makes a manager, manager can't read or write (update or delete)",

		SupportedMutationTypes: immutable.Some(
			[]testUtils.MutationType{
				// GQL mutation will return no error when wrong identity is used with gql (only for update requests),
				testUtils.CollectionNamedMutationType,
				testUtils.CollectionSaveMutationType,
			}),

		Actions: []any{
			testUtils.AddDACPolicy{

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

			testUtils.AddDACActorRelationship{ // Make admin / manager
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedExistence: false,
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(2), // Manager can not read

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

			testUtils.UpdateDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2), // Manager can not update.

				DocID: 0,

				Doc: `
					{
						"name": "Shahzad Lone"
					}
				`,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2), // Manager can not delete.

				DocID: 0,

				ExpectedError: "document not found or not authorized to access",
			},

			testUtils.AddDACActorRelationship{ // Manager can manage only.
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(3),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_CantMakeRelationshipIfNotOwnerOrManager_Error(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, cant make relation if identity doesn't own or manage object, return error",

		Actions: []any{
			testUtils.AddDACPolicy{

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

			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(2), // This identity can not manage as not an admin yet

				TargetIdentity: testUtils.ClientIdentity(3),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,

				ExpectedError: "failed to add document actor relationship with acp",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
