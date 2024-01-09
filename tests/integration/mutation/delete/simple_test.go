// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package delete

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationDeletion_WithoutSubSelection(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Delete without sub-selection, should give error.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					delete_User
				}`,
				ExpectedError: "Field \"delete_User\" of type \"[User]\" must have a sub selection.",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationDeletion_WithoutSubSelectionFields(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Delete without sub-selection fields, should give error.",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `mutation {
					delete_User{

					}
				}`,
				ExpectedError: "Syntax Error GraphQL request (2:17) Unexpected empty IN {}",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
