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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_ManagerRevokesReadAccess_OtherActorCanNoLongerRead(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, manager revokes read access, other actor that can read before no longer read.",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: testUtils.UserIdentity(1),

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
				Identity: testUtils.UserIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.AddDocActorRelationship{ // Owner makes admin / manager
				RequestorIdentity: testUtils.UserIdentity(1),

				TargetIdentity: testUtils.UserIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedExistence: false,
			},

			testUtils.AddDocActorRelationship{ // Owner gives an actor read access
				RequestorIdentity: testUtils.UserIdentity(1),

				TargetIdentity: testUtils.UserIdentity(3),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			testUtils.Request{
				Identity: testUtils.UserIdentity(3), // The other actor can read

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
							"name": "Shahzad",
							"age":  int64(28),
						},
					},
				},
			},

			testUtils.DeleteDocActorRelationship{ // Admin revokes access of the other actor that could read.
				RequestorIdentity: testUtils.UserIdentity(2),

				TargetIdentity: testUtils.UserIdentity(3),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedRecordFound: true,
			},

			// The other actor can no longer read.
			testUtils.Request{
				Identity: testUtils.UserIdentity(3),

				Request: `
					query {
						Users {
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

func TestACP_OwnerRevokesManagersAccess_ManagerCanNoLongerManageOthers(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, owner revokes manager's access, manager can not longer manage others.",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: testUtils.UserIdentity(1),

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
				Identity: testUtils.UserIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.AddDocActorRelationship{ // Owner makes admin / manager
				RequestorIdentity: testUtils.UserIdentity(1),

				TargetIdentity: testUtils.UserIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedExistence: false,
			},

			testUtils.AddDocActorRelationship{ // Manager gives an actor read access
				RequestorIdentity: testUtils.UserIdentity(2),

				TargetIdentity: testUtils.UserIdentity(3),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			testUtils.Request{
				Identity: testUtils.UserIdentity(3), // The other actor can read

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
							"name": "Shahzad",
							"age":  int64(28),
						},
					},
				},
			},

			testUtils.DeleteDocActorRelationship{ // Admin revokes access of the admin.
				RequestorIdentity: testUtils.UserIdentity(1),

				TargetIdentity: testUtils.UserIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedRecordFound: true,
			},

			testUtils.AddDocActorRelationship{ // Manager can no longer grant read access.
				RequestorIdentity: testUtils.UserIdentity(2),

				TargetIdentity: testUtils.UserIdentity(4), // This identity has no access previously.

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedError: "failed to add document actor relationship with acp",
			},

			testUtils.Request{
				Identity: testUtils.UserIdentity(4), // The other actor can ofcourse still not read.

				Request: `
					query {
						Users {
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

func TestACP_AdminTriesToRevokeOwnersAccess_NotAllowedError(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, admin tries to revoke owner's access, not allowed error.",

		Actions: []any{
			testUtils.AddPolicy{

				Identity: testUtils.UserIdentity(1),

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
				Identity: testUtils.UserIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.AddDocActorRelationship{ // Owner makes admin / manager
				RequestorIdentity: testUtils.UserIdentity(1),

				TargetIdentity: testUtils.UserIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedExistence: false,
			},

			testUtils.DeleteDocActorRelationship{ // Admin tries to revoke owners `owner` relation.
				RequestorIdentity: testUtils.UserIdentity(2),

				TargetIdentity: testUtils.UserIdentity(1),

				CollectionID: 0,

				DocID: 0,

				Relation: "owner",

				ExpectedError: "cannot delete an owner relationship",
			},

			testUtils.DeleteDocActorRelationship{ // Owner can still perform owner operations, like restrict admin.
				RequestorIdentity: testUtils.UserIdentity(1),

				TargetIdentity: testUtils.UserIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedRecordFound: true,
			},

			testUtils.Request{
				Identity: testUtils.UserIdentity(1), // The owner can still read

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
							"name": "Shahzad",
							"age":  int64(28),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
