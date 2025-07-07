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
	clientCEnvName    = "DEFRA_CLIENT_C"
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
	// CClientType enables running the test suite using
	// the C implementation of the client.TxnStore interface.
	CClientType ClientType = "c"
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
