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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_DeleteDocActorRelationshipWithDummyRelationDefinedOnPolicy_NothingChanges(t *testing.T) {
	test := testUtils.TestCase{

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

			&action.AddSchema{
				Schema: `
						type Users @policy(
							id: "{{.Policy0}}",
							resource: "users"
						) {
							name: String
							age: Int
						}
					`,
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
					"Users": []map[string]any{}, // Can't see the documents
				},
			},

			testUtils.DeleteDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "dummy", // Doesn't mean anything to the database.

				ExpectedRecordFound: false,
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(2), // This identity can still not read.

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
					"Users": []map[string]any{}, // Can't see the documents
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestACP_DeleteDocActorRelationshipWithDummyRelationNotDefinedOnPolicy_Error(t *testing.T) {
	test := testUtils.TestCase{

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

			&action.AddSchema{
				Schema: `
						type Users @policy(
							id: "{{.Policy0}}",
							resource: "users"
						) {
							name: String
							age: Int
						}
					`,
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
					"Users": []map[string]any{}, // Can't see the documents
				},
			},

			testUtils.DeleteDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "NotOnPolicy", // Doesn't mean anything to the database and not on policy either.

				ExpectedError: "failed to delete document actor relationship with acp",
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(2), // This identity can still not read.

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
					"Users": []map[string]any{}, // Can't see the documents
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
