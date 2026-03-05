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

package backup

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var userCollection = (`
	type User {
		name: String
		age: Int
		book: Book
	}

	type Book {
		name: String
		author: User @primary
	}
`)

func executeTestCase(t *testing.T, test testUtils.TestCase) {
	testUtils.ExecuteTestCase(
		t,
		testUtils.TestCase{
			Actions: append(
				[]any{
					&action.AddCollection{
						SDL: userCollection,
					},
				},
				test.Actions...,
			),
		},
	)
}
