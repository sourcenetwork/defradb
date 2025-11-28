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

// CollectionDescribe executes the `client collection describe` command and requires that the returned
// result matches the expected value.
type CollectionDescribe struct {
	stateful
	augmented

	// The expected results.
	//
	// Each item will be compared individually, if ID or RootID on the
	// expected item are default they will not be compared with the actual.
	//
	// Assertions on Indexes and Sources will not distinguish between nil and empty (in order
	// to allow their ommission in most cases).
	Expected []client.CollectionVersion
}

var _ Action = (*CollectionDescribe)(nil)

func (a *CollectionDescribe) Execute() {
	args := []string{"client", "collection", "describe"}
	args = append(args, a.AdditionalArgs...)
	args = a.AppendDirections(args)

	result, err := executeJson[[]client.CollectionVersion](a.s.Ctx, args)
	require.NoError(a.s.T, err)

	require.Equal(a.s.T, len(a.Expected), len(result))

	for i, expected := range a.Expected {
		actual := result[i]

		if expected.CollectionID != "" {
			require.Equal(a.s.T, expected.CollectionID, actual.CollectionID)
		}
		if expected.VersionID != "" {
			require.Equal(a.s.T, expected.VersionID, actual.VersionID)
		}

		require.Equal(a.s.T, expected.Name, actual.Name)
		require.Equal(a.s.T, expected.IsMaterialized, actual.IsMaterialized)
		require.Equal(a.s.T, expected.IsBranchable, actual.IsBranchable)

		if expected.Indexes != nil || len(actual.Indexes) != 0 {
			// Dont bother asserting this if the expected is nil and the actual is nil/empty.
			// This is to save each test action from having to bother declaring an empty slice (if there are no indexes)
			require.Equal(a.s.T, expected.Indexes, actual.Indexes)
		}

		require.Equal(a.s.T, expected.PreviousVersion.HasValue(), actual.PreviousVersion.HasValue())
		if expected.PreviousVersion.HasValue() {
			require.Equal(
				a.s.T,
				expected.PreviousVersion.Value().SourceCollectionID,
				actual.PreviousVersion.Value().SourceCollectionID,
			)
			require.Equal(
				a.s.T,
				expected.PreviousVersion.Value().Transform.HasValue(),
				actual.PreviousVersion.Value().Transform.HasValue(),
			)

			if expected.PreviousVersion.Value().Transform.HasValue() {
				// Dont bother asserting this by default, the transform object is too complex to bother with in most cases.
				require.Equal(
					a.s.T,
					expected.PreviousVersion.Value().Transform.Value(),
					actual.PreviousVersion.Value().Transform.Value(),
				)
			}
		}

		if expected.Query.HasValue() {
			// Dont bother asserting this by default, the query object is too complex to bother with in most cases.
			require.Equal(a.s.T, expected.Query, actual.Query)
		}

		if expected.Fields != nil {
			require.Equal(a.s.T, expected.Fields, actual.Fields)
		}

		if expected.VectorEmbeddings != nil {
			require.Equal(a.s.T, expected.VectorEmbeddings, actual.VectorEmbeddings)
		}
	}
}
