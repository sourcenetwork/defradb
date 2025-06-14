// Copyright 2025 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/clients/js"
)

func init() {
	goClient = false
	httpClient = false
	cliClient = false
	jsClient = true
	// JavaScript networking stack is managed externally
	skipNetworkTests = true
	// Backup API is not suitable for browser environments
	skipBackupTests = true
}

// setupClient returns the client implementation for the current
// testing state. The client type on the test state is used to
// select the client implementation to use.
func setupClient(_ *state, node *node.Node) (impl clients.Client, err error) {
	return js.NewWrapper(node)
}
