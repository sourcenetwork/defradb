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

package one_to_one_to_one

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func execute(t *testing.T, test testUtils.TestCase) {
	testUtils.ExecuteTestCase(
		t,
		testUtils.TestCase{
			SupportedDatabaseTypes: test.SupportedDatabaseTypes,
			Actions: append(
				[]any{
					&action.AddCollection{
						SDL: `
							type Book {
								name: String
								rating: Float
								author: Author
								publisher: Publisher @primary
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
