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
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/tests/state"
)

// Asserts as to whether an error has been raised as expected (or not). If an expected
// error has been raised it will return true, returns false in all other cases.
func assertError(t testing.TB, err error, expectedError string) bool {
	if err == nil {
		return false
	}

	if expectedError == "" {
		require.NoError(t, err)
		return false
	} else {
		if !strings.Contains(err.Error(), expectedError) {
			// Must be require instead of assert, otherwise will show a fake "error not raised".
			require.ErrorIs(t, err, errors.New(expectedError))
			return false
		}
		return true
	}
}

func assertExpectedErrorRaised(t testing.TB, expectedError string, wasRaised bool) {
	if expectedError != "" && !wasRaised {
		assert.Fail(t, "Expected an error however none was raised.")
	}
}

func assertCollectionVersions(
	s *state.State,
	expected []client.CollectionVersion,
	actual []client.CollectionVersion,
) {
	require.Equal(s.T, len(expected), len(actual))

	for i, expected := range expected {
		actual := actual[i]
		require.Equal(s.T, expected.Name, actual.Name)

		if expected.CollectionSet.HasValue() {
			require.Equal(s.T, expected.CollectionSet.Value().CollectionSetID, actual.CollectionSet.Value().CollectionSetID)
			require.Equal(s.T, expected.CollectionSet.Value().RelativeID, actual.CollectionSet.Value().RelativeID)
		}

		if expected.VersionID != "" {
			require.Equal(s.T, expected.VersionID, actual.VersionID)
		}
		if expected.CollectionID != "" {
			require.Equal(s.T, expected.CollectionID, actual.CollectionID)
		}

		require.Equal(s.T, expected.IsMaterialized, actual.IsMaterialized)
		require.Equal(s.T, expected.IsBranchable, actual.IsBranchable)
		require.Equal(s.T, expected.IsActive, actual.IsActive)

		if expected.Indexes != nil || len(actual.Indexes) != 0 {
			// Dont bother asserting this if the expected is nil and the actual is nil/empty.
			// This is to save each test action from having to bother declaring an empty slice (if there are no indexes)
			require.Equal(s.T, expected.Indexes, actual.Indexes)
		}

		if expected.Sources != nil {
			// Dont bother asserting this if the expected is nil and the actual is nil/empty.
			// This is to save each test action from having to bother declaring an empty slice (if there are no sources)
			require.Equal(s.T, expected.Sources, actual.Sources)
		}

		if expected.Fields != nil {
			require.Equal(s.T, len(expected.Fields), len(actual.Fields))
			for i := range expected.Fields {
				expectedField := expected.Fields[i]
				actualField := actual.Fields[i]

				require.Equal(s.T, expectedField.Name, actualField.Name)
				if expectedField.FieldID != "" {
					require.Equal(s.T, expectedField.FieldID, actualField.FieldID)
				}
				require.Equal(s.T, expectedField.IsPrimary, actualField.IsPrimary)
				require.Equal(s.T, expectedField.Kind, actualField.Kind)
				require.Equal(s.T, expectedField.Typ, actualField.Typ)
				require.Equal(s.T, expectedField.DefaultValue, actualField.DefaultValue)
				require.Equal(s.T, expectedField.RelationName, actualField.RelationName)
				require.Equal(s.T, expectedField.Size, actualField.Size)
			}
		}

		if expected.VectorEmbeddings != nil {
			require.Equal(s.T, expected.VectorEmbeddings, actual.VectorEmbeddings)
		}
	}
}
