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

package state

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
