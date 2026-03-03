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
	"os"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/state"
)

const (
	lensTypeEnvName = "DEFRA_LENS_TYPE"
)

var (
	lensType options.NodeLensRuntimeType
)

func init() {
	lensType = options.NodeLensRuntimeType(os.Getenv(lensTypeEnvName))
}

// ConfigureMigration is a test action which will configure a Lens migration using the
// provided configuration.
type ConfigureMigration struct {
	// NodeID is the node ID (index) of the node in which to configure the migration.
	NodeID immutable.Option[int]

	// Used to identify the transaction for this to run against. Optional.
	TransactionID immutable.Option[int]

	// Identity is the identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

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

func configureMigration(
	s *state.State,
	action ConfigureMigration,
) {
	var lensID string

	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		migrationOpts := options.SetMigration()
		identOption := getIdentityForRequestSpecificToNode(s, action.Identity, nodeID)
		if identOption.HasValue() {
			migrationOpts.SetIdentity(identOption.Value())
		}

		txn := getTransaction(s, node.Client, action.TransactionID, action.ExpectedError)
		ctx := db.InitContext(s.Ctx, txn)
		var err error
		lensID, err = node.SetMigration(ctx, action.LensConfig, migrationOpts)
		expectedErrorRaised := AssertError(s.T, err, action.ExpectedError)

		assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)
	}

	s.LensIDs = append(s.LensIDs, lensID)

	// After setting migration the collection's Version.Previous.Value().Transform should be set.
	// that's why we need to refresh collections, so that the in-memory collection versions are updated.
	// Originally was added for [NewIndex] to be able to index docs with migrated values.
	refreshCollections(s, action.TransactionID, immutable.None[state.Identity]())
}
