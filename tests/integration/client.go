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
	"fmt"
	"os"
	"strconv"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/net"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/clients/cli"
	"github.com/sourcenetwork/defradb/tests/clients/http"
)

const (
	clientGoEnvName   = "DEFRA_CLIENT_GO"
	clientHttpEnvName = "DEFRA_CLIENT_HTTP"
	clientCliEnvName  = "DEFRA_CLIENT_CLI"
)

type ClientType string

const (
	// goClientType enables running the test suite using
	// the go implementation of the client.DB interface.
	GoClientType ClientType = "go"
	// httpClientType enables running the test suite using
	// the http implementation of the client.DB interface.
	HTTPClientType ClientType = "http"
	// cliClientType enables running the test suite using
	// the cli implementation of the client.DB interface.
	CLIClientType ClientType = "cli"
)

var (
	httpClient bool
	goClient   bool
	cliClient  bool
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

// setupClient returns the client implementation for the current
// testing state. The client type on the test state is used to
// select the client implementation to use.
func setupClient(s *state, node *node.Node) (impl clients.Client, err error) {
	switch s.clientType {
	case HTTPClientType:
		impl, err = http.NewWrapper(node)

	case CLIClientType:
		impl, err = cli.NewWrapper(node)

	case GoClientType:
		impl = newGoClientWrapper(node)

	default:
		err = fmt.Errorf("invalid client type: %v", s.dbt)
	}

	if err != nil {
		return nil, err
	}
	return
}

type goClientWrapper struct {
	client.DB
	peer *net.Peer
}

func newGoClientWrapper(n *node.Node) *goClientWrapper {
	return &goClientWrapper{
		DB:   n.DB,
		peer: n.Peer,
	}
}

func (w *goClientWrapper) Bootstrap(addrs []peer.AddrInfo) {
	if w.peer != nil {
		w.peer.Bootstrap(addrs)
	}
}

func (w *goClientWrapper) Close() {
	if w.peer != nil {
		w.peer.Close()
	}
	w.DB.Close()
}
