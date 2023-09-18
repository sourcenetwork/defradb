// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

// ConfigureMigration is a test action which will configure a Lens migration using the
// provided configuration.
type ConfigureMigration struct {
	// NodeID is the node ID (index) of the node in which to configure the migration.
	NodeID immutable.Option[int]

	// Used to identify the transaction for this to run against. Optional.
	TransactionID immutable.Option[int]

	// The configuration to use.
	//
	// Paths to WASM Lens modules may be found in: github.com/sourcenetwork/defradb/tests/lenses
	client.LensConfig

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// GetMigrations is a test action which will fetch and assert on the results of calling
// `LensRegistry().Config()`.
type GetMigrations struct {
	// NodeID is the node ID (index) of the node in which to configure the migration.
	NodeID immutable.Option[int]

	// Used to identify the transaction for this to run against. Optional.
	TransactionID immutable.Option[int]

	// The expected configuration.
	ExpectedResults []client.LensConfig
}

func configureMigration(
	s *state,
	action ConfigureMigration,
) {
	for _, node := range getNodes(action.NodeID, s.nodes) {
		db := getStore(s, node.DB, action.TransactionID, action.ExpectedError)

		err := db.SetMigration(s.ctx, action.LensConfig)
		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)

		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
	}
}

func getMigrations(
	s *state,
	action GetMigrations,
) {
	for _, node := range getNodes(action.NodeID, s.nodes) {
		db := getStore(s, node.DB, action.TransactionID, "")

		configs, err := db.LensRegistry().Config(s.ctx)
		require.NoError(s.t, err)
		require.Equal(s.t, len(configs), len(action.ExpectedResults))

		// The order of the results is not deterministic, so do not assert on the element
		for _, expected := range action.ExpectedResults {
			var actual client.LensConfig
			var actualFound bool

			for _, config := range configs {
				if config.SourceSchemaVersionID != expected.SourceSchemaVersionID {
					continue
				}
				if config.DestinationSchemaVersionID != expected.DestinationSchemaVersionID {
					continue
				}
				actual = config
				actualFound = true
			}

			require.True(s.t, actualFound, "matching lens config not found")
			require.Equal(s.t, len(expected.Lenses), len(actual.Lenses))

			for j, actualLens := range actual.Lenses {
				expectedLens := expected.Lenses[j]

				assert.Equal(s.t, expectedLens.Inverse, actualLens.Inverse)
				assert.Equal(s.t, expectedLens.Path, actualLens.Path)

				assertResultsEqual(s.t, s.clientType, expectedLens.Arguments, actualLens.Arguments)
			}
		}
	}
}
