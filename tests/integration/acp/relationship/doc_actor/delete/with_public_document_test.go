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

func TestACP_DeleteDocActorRelationshipWithPublicDocument_CanAlreadyAccess_Error(t *testing.T) {
	expectedPolicyID := "fc56b7509c20ac8ce682b3b9b4fdaad868a9c70dda6ec16720298be64f16e9a4"

	test := testUtils.TestCase{

		Description: "Test acp, delete doc actor relationship on a public document, return error",

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

			testUtils.CreateDoc{ // Note: Is a public document (without an identity).
				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			testUtils.Request{
				Identity: immutable.Some(2), // Can read as it is a public document

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

			testUtils.DeleteDocActorRelationship{
				RequestorIdentity: 1,

				TargetIdentity: 2,

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedError: "failed to delete document actor relationship with acp",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
