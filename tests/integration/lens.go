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
	"context"
	"testing"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/net"
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
	ctx context.Context,
	t *testing.T,
	nodes []*net.Node,
	txnsPointer *[]datastore.Txn,
	testCase TestCase,
	action ConfigureMigration,
) {
	for _, node := range getNodes(action.NodeID, nodes) {
		db := getStore(ctx, t, testCase.Description, node.DB, txnsPointer, action.TransactionID, action.ExpectedError)

		err := db.SetMigration(ctx, action.LensConfig)
		expectedErrorRaised := AssertError(t, testCase.Description, err, action.ExpectedError)

		assertExpectedErrorRaised(t, testCase.Description, action.ExpectedError, expectedErrorRaised)
	}
}

func getMigrations(
	ctx context.Context,
	t *testing.T,
	nodes []*net.Node,
	txnsPointer *[]datastore.Txn,
	testCase TestCase,
	action GetMigrations,
) {
	for _, node := range getNodes(action.NodeID, nodes) {
		db := getStore(ctx, t, testCase.Description, node.DB, txnsPointer, action.TransactionID, "")

		configs := db.LensRegistry().Config()

		// The order of the results is not deterministic, so do not assert on the element
		// locations.
		assert.ElementsMatch(t, configs, action.ExpectedResults)
	}
}
