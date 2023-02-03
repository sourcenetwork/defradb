// Copyright 2023 Democratized Data Foundation
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
	type users {
		Name: String
		LikedIndexes: [Boolean!]
		IndexLikesDislikes: [Boolean]
		FavouriteIntegers: [Int!]
		TestScores: [Int]
		FavouriteFloats: [Float!]
		PageRatings: [Float]
		PreferredStrings: [String!]
		PageHeaders: [String]
	}
`)

func executeTestCase(t *testing.T, test testUtils.RequestTestCase) {
	testUtils.ExecuteRequestTestCase(t, userCollectionGQLSchema, []string{"users"}, test)
}
