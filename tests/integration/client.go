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
	"os"
	"strconv"
)

const (
	clientGoEnvName   = "DEFRA_CLIENT_GO"
	clientHttpEnvName = "DEFRA_CLIENT_HTTP"
	clientCliEnvName  = "DEFRA_CLIENT_CLI"
	clientCEnvName    = "DEFRA_CLIENT_C"
)

var (
	httpClient bool
	goClient   bool
	cliClient  bool
	jsClient   bool
	cClient    bool
)

func init() {
	// We use environment variables instead of flags `go test ./...` throws for all packages
	// that don't have the flag defined
	httpClient, _ = strconv.ParseBool(os.Getenv(clientHttpEnvName))
	goClient, _ = strconv.ParseBool(os.Getenv(clientGoEnvName))
	cliClient, _ = strconv.ParseBool(os.Getenv(clientCliEnvName))
	cClient, _ = strconv.ParseBool(os.Getenv(clientCEnvName))

	if !goClient && !httpClient && !cliClient && !cClient {
		// Default is to test go client type.
		goClient = true
	}
}
