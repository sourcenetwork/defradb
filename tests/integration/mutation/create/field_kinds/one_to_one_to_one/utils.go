// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one_to_one

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func execute(t *testing.T, test testUtils.TestCase) {
	testUtils.ExecuteTestCase(
		t,
		testUtils.TestCase{
			Description: test.Description,
			Actions: append(
				[]any{
					testUtils.SchemaUpdate{
						Schema: `
							type Book {
								name: String
								rating: Float
								author: Author
								publisher: Publisher
							}

							type Author {
								name: String
								age: Int
								verified: Boolean
								wrote: Book @primary
							}

							type Publisher {
								name: String
								address: String
								published: Book
							}
						`,
					},
				},
				test.Actions...,
			),
		},
	)
}
