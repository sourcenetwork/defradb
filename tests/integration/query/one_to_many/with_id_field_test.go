// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_many

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"

	"github.com/sourcenetwork/immutable"
)

// This documents unwanted behaviour, see https://github.com/sourcenetwork/defradb/issues/1520
func TestQueryOneToManyWithIdFieldOnPrimary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "One-to-many relation primary direction, id field with name clash on primary side",
		SupportedMutationTypes: immutable.Some([]testUtils.MutationType{
			// GQL mutation will return a different error
			// when field types do not match
			testUtils.CollectionNamedMutationType,
			testUtils.CollectionSaveMutationType,
		}),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Book {
						name: String
						author_id: Int
						author: Author
					}

					type Author {
						name: String
						published: [Book]
					}
				`,
				ExpectedError: "relational id field of invalid kind. Field: author_id, Expected: ID, Actual: Int",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
