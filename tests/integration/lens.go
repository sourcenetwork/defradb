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
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/node"
)

const (
	lensTypeEnvName = "DEFRA_LENS_TYPE"
)

var (
	lensType node.LensRuntimeType
)

func init() {
	lensType = node.LensRuntimeType(os.Getenv(lensTypeEnvName))
}

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

func configureMigration(
	s *state,
	action ConfigureMigration,
) {
	_, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for _, node := range nodes {
		txn := getTransaction(s, node.DB, action.TransactionID, action.ExpectedError)
		ctx := db.SetContextTxn(s.ctx, txn)

		err := node.DB.SetMigration(ctx, action.LensConfig)
		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)

		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
	}
}
