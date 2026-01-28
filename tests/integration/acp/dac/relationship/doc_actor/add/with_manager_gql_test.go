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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestACP_OwnerMakesAManagerThatGivesItSelfReadAndWriteAccess_GQL_ManagerCanReadAndWrite(t *testing.T) {
	test := testUtils.TestCase{

		SupportedMutationTypes: immutable.Some(
			[]state.MutationType{
				// GQL mutation will return no error when wrong identity is used (only for update requests),
				// so test that separately.
				state.GQLRequestMutationType,
			},
		),

		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: A Policy
name: Test Policy
resources:
- name: users
  permissions:
  - expr: deleter
    name: delete
  - expr: dummy
    name: nothing
  - expr: reader + updater + deleter
    name: read
  - expr: updater
    name: update
  relations:
  - manages:
    - reader
    - updater
    - deleter
    name: admin
    types:
    - actor
  - name: deleter
    types:
    - actor
  - name: dummy
    types:
    - actor
  - name: reader
    types:
    - actor
  - name: updater
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

			&action.CreateDoc{
				Identity: testUtils.ClientIdentity(1),

				CollectionID: 0,

				Doc: `
					{
						"name": "Shahzad",
						"age": 28
					}
				`,
			},

			&action.Request{
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

				SkipLocalUpdateEvent: true,
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

			testUtils.AddDACActorRelationship{ // Manager makes itself a updater.
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

			&action.Request{
				Identity: testUtils.ClientIdentity(2), // Manager can read now

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

			testUtils.DeleteDoc{
				CollectionID: 0,

				Identity: testUtils.ClientIdentity(2), // Manager still can not delete yet.

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

				Identity: testUtils.ClientIdentity(2), // Manager can delete now.

				DocID: 0,
			},

			&action.Request{
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

func TestACP_OwnerMakesManagerButManagerCanNotPerformOperations_GQL_ManagerCantReadOrWrite(t *testing.T) {
	test := testUtils.TestCase{

		SupportedMutationTypes: immutable.Some(
			[]state.MutationType{
				// GQL mutation will return no error when wrong identity is used (only for update requests),
				// so test that separately.
				state.GQLRequestMutationType,
			},
		),

		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: A Policy
name: Test Policy
resources:
- name: users
  permissions:
  - expr: deleter
    name: delete
  - expr: dummy
    name: nothing
  - expr: reader + updater + deleter
    name: read
  - expr: updater
    name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: deleter
    types:
    - actor
  - name: dummy
    types:
    - actor
  - name: reader
    types:
    - actor
  - name: updater
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

			&action.CreateDoc{
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

			&action.Request{
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

				SkipLocalUpdateEvent: true,
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

func TestACP_ManagerAddsRelationshipWithRelationItDoesNotManageAccordingToPolicy_GQL_Error(t *testing.T) {
	test := testUtils.TestCase{

		SupportedMutationTypes: immutable.Some(
			[]state.MutationType{
				// GQL mutation will return no error when wrong identity is used (only for update requests),
				// so test that separately.
				state.GQLRequestMutationType,
			},
		),

		Actions: []any{
			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
description: A Policy
name: Test Policy
resources:
- name: users
  permissions:
  - expr: deleter
    name: delete
  - expr: dummy
    name: nothing
  - expr: reader + updater + deleter
    name: read
  - expr: updater
    name: update
  relations:
  - manages:
    - reader
    name: admin
    types:
    - actor
  - name: deleter
    types:
    - actor
  - name: dummy
    types:
    - actor
  - name: reader
    types:
    - actor
  - name: updater
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

			&action.CreateDoc{
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

			testUtils.AddDACActorRelationship{ // Admin tries to make another actor an updater
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(3),

				CollectionID: 0,

				DocID: 0,

				Relation: "updater",

				ExpectedError: "UNAUTHORIZED",
			},

			testUtils.AddDACActorRelationship{ // Admin tries to make another actor a deleter
				RequestorIdentity: testUtils.ClientIdentity(2),

				TargetIdentity: testUtils.ClientIdentity(3),

				CollectionID: 0,

				DocID: 0,

				Relation: "deleter",

				ExpectedError: "UNAUTHORIZED",
			},

			&action.Request{
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

				SkipLocalUpdateEvent: true,
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
