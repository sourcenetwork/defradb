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

//go:build javaclient

package tests

import (
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
	javaclient "github.com/sourcenetwork/defradb/tests/clients/java"
)

func setupJavaClient(nodeObj *node.Node) (clients.Client, error) {
	return javaclient.NewWrapper(nodeObj)
}
