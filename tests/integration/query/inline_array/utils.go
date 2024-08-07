// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package inline_array

import (
	"testing"

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
	testUtils.ExecuteTestCase(
		t,
		testUtils.TestCase{
			Description: test.Description,
			Actions: append(
				[]any{
					testUtils.SchemaUpdate{
						Schema: userCollectionGQLSchema,
					},
				},
				test.Actions...,
			),
		},
	)
}
