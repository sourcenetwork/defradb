// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package tests

import (
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/clients/js"
	"github.com/sourcenetwork/defradb/tests/state"
)

func init() {
	goClient = false
	httpClient = false
	cliClient = false
	cClient = false
	jsClient = true
	// JavaScript networking stack is managed externally
	skipNetworkTests = true
	// Backup API is not suitable for browser environments
	skipBackupTests = true
}

// setupClient returns the client implementation for the current
// testing state. The client type on the test state is used to
// select the client implementation to use.
func setupClient(_ *state.State, node *node.Node) (impl clients.Client, err error) {
	return js.NewWrapper(node)
}
