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
	"strconv"
)

const (
	clientGoEnvName   = "DEFRA_CLIENT_GO"
	clientHttpEnvName = "DEFRA_CLIENT_HTTP"
	clientCliEnvName  = "DEFRA_CLIENT_CLI"
)

type ClientType string

const (
	// goClientType enables running the test suite using
	// the go implementation of the client.TxnStore interface.
	GoClientType ClientType = "go"
	// httpClientType enables running the test suite using
	// the http implementation of the client.TxnStore interface.
	HTTPClientType ClientType = "http"
	// cliClientType enables running the test suite using
	// the cli implementation of the client.TxnStore interface.
	CLIClientType ClientType = "cli"
	// JSClientType enables running the test suite using
	// the JS implementation of the client.TxnStore interface.
	JSClientType ClientType = "js"
)

var (
	httpClient bool
	goClient   bool
	cliClient  bool
	jsClient   bool
)

func init() {
	// We use environment variables instead of flags `go test ./...` throws for all packages
	// that don't have the flag defined
	httpClient, _ = strconv.ParseBool(os.Getenv(clientHttpEnvName))
	goClient, _ = strconv.ParseBool(os.Getenv(clientGoEnvName))
	cliClient, _ = strconv.ParseBool(os.Getenv(clientCliEnvName))

	if !goClient && !httpClient && !cliClient {
		// Default is to test go client type.
		goClient = true
	}
}
