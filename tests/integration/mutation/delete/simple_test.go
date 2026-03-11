// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package delete

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationDeletion_WithoutSubSelection(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.Request{
				Request: `mutation {
					delete_User{

					}
				}`,
				ExpectedError: "Syntax Error GraphQL",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
