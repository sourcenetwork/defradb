// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"context"
	"fmt"
	gohttp "net/http"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/http"
)

func (n *Node) startAPI(ctx context.Context) error {
	if n.config.disableAPI {
		return nil
	}
	handler, err := http.NewHandler(n.DB)
	if err != nil {
		return err
	}
	n.server, err = http.NewServer(handler, filterOptions[http.ServerOpt](n.options)...)
	if err != nil {
		return err
	}
	err = n.server.SetListener()
	if err != nil {
		return err
	}
	log.InfoContext(ctx,
		fmt.Sprintf("Providing HTTP API at %s PlaygroundEnabled=%t", n.server.Address(), http.PlaygroundEnabled))
	log.InfoContext(ctx, fmt.Sprintf("Providing GraphQL endpoint at %s/api/v0/graphql", n.server.Address()))
	go func() {
		if err := n.server.Serve(); err != nil && !errors.Is(err, gohttp.ErrServerClosed) {
			log.ErrorContextE(ctx, "HTTP server stopped", err)
		}
	}()
	n.APIURL = n.server.Address()
	// Check that the server is ready before returning. We do this to ensure that
	// subsequent operation will behave as expected.
	c, err := http.NewClient(n.APIURL)
	if err != nil {
		return err
	}
	err = c.HealthCheck(ctx)
	if err != nil {
		return err
	}
	return nil
}
