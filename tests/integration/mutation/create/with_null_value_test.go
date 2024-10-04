// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package create

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationCreate_WithOmittedValueAndExplicitNullValue(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// Collection.Save would treat the second create as an update, and so
			// is excluded from this test.
			testUtils.CollectionNamedMutationType,
			testUtils.GQLRequestMutationType,
		}),
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
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"age": null
				}`,
				ExpectedError: "a document with the given ID already exist",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
