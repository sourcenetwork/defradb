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

package add

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestMutationAdd_WithOmittedValueAndExplicitNullValue(t *testing.T) {
	test := testUtils.TestCase{
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			// Collection.Save would treat the second create as an update, and so
			// is excluded from this test.
			state.CollectionNamedMutationType,
			state.GQLRequestMutationType,
		}),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						age: Int
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.AddDoc{
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
