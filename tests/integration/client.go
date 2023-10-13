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

	"github.com/sourcenetwork/defradb/net"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/clients/cli"
	"github.com/sourcenetwork/defradb/tests/clients/http"
)

// setupClient returns the client implementation for the current
// testing state. The client type on the test state is used to
// select the client implementation to use.
func setupClient(s *state, node *net.Node) (impl clients.Client, err error) {
	switch s.clientType {
	case httpClientType:
		impl, err = http.NewWrapper(node)

	case cliClientType:
		impl = cli.NewWrapper(node)

	case goClientType:
		impl = node

	default:
		err = fmt.Errorf("invalid client type: %v", s.dbt)
	}

	if err != nil {
		return nil, err
	}
	return
}
