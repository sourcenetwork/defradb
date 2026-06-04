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

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestEncryptFieldsForAddMutation(t *testing.T) {
	test := testUtils.TestCase{

		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						age:  Int
					}
				`,
			},
			testUtils.IntrospectionRequest{
				Request: `
				{
				  __type(name: "UserField") {
				    name
				    kind
				    enumValues {
				      name
				    }
				  }
				}
				`,
				ContainsData: map[string]any{
					"__type": map[string]any{
						"kind": "ENUM",
						"name": "UserField",
						"enumValues": []any{
							map[string]any{"name": "name"},
							map[string]any{"name": "age"},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
