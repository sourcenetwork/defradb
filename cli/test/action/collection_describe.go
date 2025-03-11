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
	argmented

	// The expected results.
	//
	// Each item will be compared individually, if ID, RootID or SchemaVersionID on the
	// expected item are default they will not be compared with the actual.
	//
	// Assertions on Indexes and Sources will not distinguish between nil and empty (in order
	// to allow their ommission in most cases).
	Expected []client.CollectionDefinition
}

var _ Action = (*CollectionDescribe)(nil)

func (a *CollectionDescribe) Execute() {
	args := []string{"client", "collection", "describe"}
	args = append(args, a.AdditionalArgs...)
	args = a.AppendDirections(args)

	result, err := executeJson[[]client.CollectionDefinition](a.s.Ctx, args)
	require.NoError(a.s.T, err)

	require.Equal(a.s.T, len(a.Expected), len(result))

	for i, expected := range a.Expected {
		actual := result[i]

		if expected.Description.ID != 0 {
			require.Equal(a.s.T, expected.Description.ID, actual.Description.ID)
		}
		if expected.Description.RootID != 0 {
			require.Equal(a.s.T, expected.Description.RootID, actual.Description.RootID)
		}
		if expected.Description.SchemaVersionID != "" {
			require.Equal(a.s.T, expected.Description.SchemaVersionID, actual.Description.SchemaVersionID)
		}

		require.Equal(a.s.T, expected.Description.Name, actual.Description.Name)
		require.Equal(a.s.T, expected.Description.IsMaterialized, actual.Description.IsMaterialized)
		require.Equal(a.s.T, expected.Description.IsBranchable, actual.Description.IsBranchable)

		if expected.Description.Indexes != nil || len(actual.Description.Indexes) != 0 {
			// Dont bother asserting this if the expected is nil and the actual is nil/empty.
			// This is to save each test action from having to bother declaring an empty slice (if there are no indexes)
			require.Equal(a.s.T, expected.Description.Indexes, actual.Description.Indexes)
		}

		if expected.Description.Sources != nil {
			// Dont bother asserting this if the expected is nil and the actual is nil/empty.
			// This is to save each test action from having to bother declaring an empty slice (if there are no sources)
			require.Equal(a.s.T, expected.Description.Sources, actual.Description.Sources)
		}

		if expected.Description.Fields != nil {
			require.Equal(a.s.T, expected.Description.Fields, actual.Description.Fields)
		}

		if expected.Description.VectorEmbeddings != nil {
			require.Equal(a.s.T, expected.Description.VectorEmbeddings, actual.Description.VectorEmbeddings)
		}
	}
}
