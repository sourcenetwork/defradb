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

func TestSchemaOneOne_NoPrimary_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						dog: Dog
					}
					type Dog {
						name: String
						owner: User
					}
				`,
				ExpectedError: "primary side of relation not defined. RelationName: dog_user",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestSchemaOneOne_TwoPrimaries_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
						dog: Dog @primary
					}
					type Dog {
						name: String
						owner: User @primary
					}
				`,
				ExpectedError: "relation can only have a single field set as primary",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
