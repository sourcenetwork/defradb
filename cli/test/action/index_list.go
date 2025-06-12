// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package action

import (
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

// IndexList executes the `client index list` command and requires that the returned
// result matches the expected value.
type IndexList struct {
	stateful
	augmented

	// The collection name to list indexes for (optional).
	// If not provided, all indexes in the database will be listed.
	Collection string

	// The expected indexes when listing for a specific collection.
	ExpectedIndexes []client.IndexDescription

	// The expected indexes when listing all indexes (map of collection name to indexes).
	ExpectedAllIndexes map[client.CollectionName][]client.IndexDescription

	// ExpectError is the expected error string. If empty, no error is expected.
	ExpectError string
}

var _ Action = (*IndexList)(nil)

func (a *IndexList) Execute() {
	args := []string{"client", "index", "list"}

	if a.Collection != "" {
		args = append(args, "--collection", a.Collection)
	}

	args = append(args, a.AdditionalArgs...)
	args = a.AppendDirections(args)

	if a.Collection != "" {
		// Listing indexes for a specific collection
		result, err := executeJson[[]client.IndexDescription](a.s.Ctx, args)

		if a.ExpectError != "" {
			require.Error(a.s.T, err)
			require.Contains(a.s.T, err.Error(), a.ExpectError)
			return
		}

		require.NoError(a.s.T, err)

		if a.ExpectedIndexes != nil {
			require.Equal(a.s.T, len(a.ExpectedIndexes), len(result))

			for i, expected := range a.ExpectedIndexes {
				actual := result[i]

				if expected.ID != 0 {
					require.Equal(a.s.T, expected.ID, actual.ID)
				}
				require.Equal(a.s.T, expected.Name, actual.Name)
				require.Equal(a.s.T, expected.Fields, actual.Fields)
				require.Equal(a.s.T, expected.Unique, actual.Unique)
			}
		}
	} else {
		// Listing all indexes
		result, err := executeJson[map[client.CollectionName][]client.IndexDescription](a.s.Ctx, args)

		if a.ExpectError != "" {
			require.Error(a.s.T, err)
			require.Contains(a.s.T, err.Error(), a.ExpectError)
			return
		}

		require.NoError(a.s.T, err)

		if a.ExpectedAllIndexes != nil {
			require.Equal(a.s.T, len(a.ExpectedAllIndexes), len(result))

			for collectionName, expectedIndexes := range a.ExpectedAllIndexes {
				actualIndexes, ok := result[collectionName]
				require.True(a.s.T, ok, "Expected collection %s not found in result", collectionName)
				require.Equal(a.s.T, len(expectedIndexes), len(actualIndexes))

				for i, expected := range expectedIndexes {
					actual := actualIndexes[i]

					if expected.ID != 0 {
						require.Equal(a.s.T, expected.ID, actual.ID)
					}
					require.Equal(a.s.T, expected.Name, actual.Name)
					require.Equal(a.s.T, expected.Fields, actual.Fields)
					require.Equal(a.s.T, expected.Unique, actual.Unique)
				}
			}
		}
	}
}
