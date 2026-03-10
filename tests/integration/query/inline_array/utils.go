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

package inline_array

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var userCollectionGQLSchema = (`
	type Users {
		name: String
		likedIndexes: [Boolean!]
		indexLikesDislikes: [Boolean]
		favouriteIntegers: [Int!]
		testScores: [Int]
		favouriteFloats: [Float!]
		pageRatings: [Float]
		preferredStrings: [String!]
		pageHeaders: [String]
	}
`)

func executeTestCase(t *testing.T, test testUtils.TestCase) {
	test.Actions = append(
		[]any{
			&action.AddCollection{
				SDL: userCollectionGQLSchema,
			},
		},
		test.Actions...,
	)
	testUtils.ExecuteTestCase(t, test)
}
