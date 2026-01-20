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
)

func TestACP_OwnerGivesOnlyReadAccessToAnotherActor_GQL_OtherActorCanReadButNotUpdate(t *testing.T) {
	test := testUtils.TestCase{

		SupportedMutationTypes: immutable.Some(
			[]testUtils.MutationType{
				// GQL mutation will return no error when wrong identity is used (only for update requests),
				// so test that separately.
				testUtils.GQLRequestMutationType,
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

			testUtils.AddDACActorRelationship{
				RequestorIdentity: testUtils.ClientIdentity(1),

				TargetIdentity: testUtils.ClientIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedExistence: false,
			},

			testUtils.Request{
				Identity: testUtils.ClientIdentity(2), // Now this identity can read.

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

			testUtils.UpdateDoc{ // But this actor still can't update.
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
