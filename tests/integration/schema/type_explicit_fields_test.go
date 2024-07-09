// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestEncryptFieldsForCreateMutation(t *testing.T) {
	test := testUtils.TestCase{

		Description: "Test that type explicit (or user-defined) fields are generated.",

		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
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
