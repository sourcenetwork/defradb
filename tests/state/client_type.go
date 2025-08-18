// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
