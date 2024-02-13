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
	"net/http/httptest"
	"testing"

	badger "github.com/sourcenetwork/badger/v4"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/db"
	"github.com/sourcenetwork/defradb/errors"
	httpapi "github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/logging"
)

var log = logging.MustNewLogger("cli")

type defraInstance struct {
	db     client.DB
	server *httptest.Server
}

func (di *defraInstance) close(ctx context.Context) {
	di.db.Close()
	di.server.Close()
}

func start(ctx context.Context) (*defraInstance, error) {
	log.FeedbackInfo(ctx, "Starting DefraDB service...")

	log.FeedbackInfo(ctx, "Building new memory store")
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)

	if err != nil {
		return nil, errors.Wrap("failed to open datastore", err)
	}

	db, err := db.NewDB(ctx, rootstore)
	if err != nil {
		return nil, errors.Wrap("failed to create a database", err)
	}

	handler, err := httpapi.NewHandler(db)
	if err != nil {
		return nil, errors.Wrap("failed to create http handler", err)
	}
	server := httptest.NewServer(handler)

	return &defraInstance{
		db:     db,
		server: server,
	}, nil
}

func startTestNode(t *testing.T) (*defraInstance, func()) {
	ctx := context.Background()
	di, err := start(ctx)
	require.NoError(t, err)
	return di, func() { di.close(ctx) }
}
