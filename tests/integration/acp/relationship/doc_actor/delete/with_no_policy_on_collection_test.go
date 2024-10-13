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
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_DeleteDocActorRelationshipWithCollectionThatHasNoPolicy_NotAllowedError(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test acp, delete doc actor relationship on a collection with no policy, not allowed error",

		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						age: Int
					}
				`,
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

			testUtils.DeleteDocActorRelationship{
				RequestorIdentity: testUtils.UserIdentity(1),

				TargetIdentity: testUtils.UserIdentity(2),

				CollectionID: 0,

				DocID: 0,

				Relation: "reader",

				ExpectedError: "operation requires ACP, but collection has no policy", // Everything is public anyway
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
