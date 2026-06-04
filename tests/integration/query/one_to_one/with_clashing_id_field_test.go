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

package one_to_one

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOneWithClashingIdFieldOnSecondary(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						_authorID: Int
						author: Author
					}

					type Author {
						name: String
						published: Book @primary
					}
				`,
				ExpectedError: "duplicate field. Name: _authorID",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneWithClashingIdFieldOnPrimary(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Book {
						name: String
						_authorID: Int
						author: Author @primary
					}

					type Author {
						name: String
						published: Book
					}
				`,
				ExpectedError: "duplicate field. Name: _authorID",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
