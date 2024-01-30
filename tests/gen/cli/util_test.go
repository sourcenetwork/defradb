// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	badger "github.com/sourcenetwork/badger/v4"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/errors"
	httpapi "github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/logging"
)

var log = logging.MustNewLogger("cli")

type defraInstance struct {
	db     client.DB
	server *httpapi.Server
}

func (di *defraInstance) close(ctx context.Context) {
	di.db.Close()
	if err := di.server.Close(); err != nil {
		log.FeedbackInfo(
			ctx,
			"The server could not be closed successfully",
			logging.NewKV("Error", err.Error()),
		)
	}
}

func start(ctx context.Context, cfg *config.Config) (*defraInstance, error) {
	log.FeedbackInfo(ctx, "Starting DefraDB service...")

	log.FeedbackInfo(ctx, "Building new memory store")
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)

	if err != nil {
		return nil, errors.Wrap("failed to open datastore", err)
	}

	db, err := db.NewDB(ctx, rootstore)
	if err != nil {
		return nil, errors.Wrap("failed to create database", err)
	}

	server, err := httpapi.NewServer(db, httpapi.WithAddress(cfg.API.Address))
	if err != nil {
		return nil, errors.Wrap("failed to create http server", err)
	}
	if err := server.Listen(ctx); err != nil {
		return nil, errors.Wrap(fmt.Sprintf("failed to listen on TCP address %v", server.Addr), err)
	}
	// save the address on the config in case the port number was set to random
	cfg.API.Address = server.AssignedAddr()
	cfg.Persist()

	// run the server in a separate goroutine
	go func(apiAddress string) {
		log.FeedbackInfo(ctx, fmt.Sprintf("Providing HTTP API at %s.", apiAddress))
		if err := server.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.FeedbackErrorE(ctx, "Failed to run the HTTP server", err)
			db.Close()
			os.Exit(1)
		}
	}(cfg.API.AddressToURL())

	return &defraInstance{
		db:     db,
		server: server,
	}, nil
}

func getTestConfig(t *testing.T) *config.Config {
	cfg := config.DefaultConfig()
	cfg.Datastore.Store = "memory"
	cfg.Net.P2PDisabled = true
	cfg.Rootdir = t.TempDir()
	cfg.Net.P2PAddresses = []string{"/ip4/127.0.0.1/tcp/0"}
	cfg.API.Address = "127.0.0.1:0"
	cfg.Persist()
	return cfg
}

func startTestNode(t *testing.T) (*config.Config, *defraInstance, func()) {
	cfg := getTestConfig(t)

	ctx := context.Background()
	di, err := start(ctx, cfg)
	require.NoError(t, err)
	return cfg, di, func() { di.close(ctx) }
}
