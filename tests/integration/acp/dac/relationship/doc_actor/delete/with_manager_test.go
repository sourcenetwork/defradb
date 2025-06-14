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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_ManagerRevokesReadAccess_OtherActorCanNoLongerRead(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, manager revokes read access, other actor that can read before no longer read.",

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

			testUtils.AddDACActorRelationship{ // Owner makes admin / manager
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedExistence: false,
			},

			testUtils.AddDACActorRelationship{ // Owner gives an actor read access
				RequestorIdentity: testUtils.ClientIdentity(1),

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

			testUtils.DeleteDACActorRelationship{ // Admin revokes access of the other actor that could read.
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(3),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedRecordFound: true,
			},

			// The other actor can no longer read.
			testUtils.Request{
				Identity: testUtils.ClientIdentity(3),

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
	test := testUtils.TestCase{

		Description: "Test acp, owner revokes manager's access, manager can not longer manage others.",

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

			testUtils.AddDACActorRelationship{ // Owner makes admin / manager
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedExistence: false,
			},

			testUtils.AddDACActorRelationship{ // Manager gives an actor read access
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

			testUtils.DeleteDACActorRelationship{ // Admin revokes access of the admin.
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedRecordFound: true,
			},

			testUtils.AddDACActorRelationship{ // Manager can no longer grant read access.
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(4), // This identity has no access previously.

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedError: "failed to add document actor relationship with acp",
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(4), // The other actor can ofcourse still not read.

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
	test := testUtils.TestCase{

		Description: "Test acp, admin tries to revoke owner's access, not allowed error.",

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

			testUtils.AddDACActorRelationship{ // Owner makes admin / manager
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedExistence: false,
			},

			testUtils.DeleteDACActorRelationship{ // Admin tries to revoke owners `owner` relation.
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(1),

				CollectionID: 0,

				DocID: 0,

				Relation: "owner",

				ExpectedError: "OPERATION_FORBIDDEN",
			},

			testUtils.DeleteDACActorRelationship{ // Owner can still perform owner operations, like restrict admin.
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "admin",

				ExpectedRecordFound: true,
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(1), // The owner can still read

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
