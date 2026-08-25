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

//go:build !javaclient

package action

import (
	"fmt"

	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
)

// setupJavaClient is a stub used whenever this binary was not built with
// -tags javaclient. Building the real implementation (tests/clients/java) 
// requires a JDK's JNI headers on CGO_CFLAGS (see that package's doc.go),
// so it's kept out of the default build.
func setupJavaClient(*node.Node) (clients.Client, error) {
	return nil, fmt.Errorf(
		"the java client requires building with -tags javaclient (see tests/clients/java/doc.go)")
}