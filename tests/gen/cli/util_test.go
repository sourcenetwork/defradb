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

	badgerds "github.com/dgraph-io/badger/v4"
	"github.com/sourcenetwork/corekv/badger"
	"github.com/sourcenetwork/corelog"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/acp/dac"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	httpapi "github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/internal/db"
)

var log = corelog.NewLogger("cli")

type DB interface {
	client.TxnStore
	Close()
}

type defraInstance struct {
	db     DB
	server *httptest.Server
}

func (di *defraInstance) close(ctx context.Context) {
	di.db.Close()
	di.server.Close()
}

func start(ctx context.Context) (*defraInstance, error) {
	log.InfoContext(ctx, "Starting DefraDB service...")

	log.InfoContext(ctx, "Building new memory store")
	rootstore, err := badger.NewDatastore("", badgerds.DefaultOptions("").WithInMemory(true))
	if err != nil {
		return nil, err
	}
	adminInfo, err := db.NewNACInfo(ctx, "", false)
	if err != nil {
		return nil, errors.Wrap("failed to setup node access control info", err)
	}
	db, err := db.NewDB(ctx, rootstore, adminInfo, dac.NoDocumentACP, nil)
	if err != nil {
		return nil, errors.Wrap("failed to create a database", err)
	}

	handler, err := httpapi.NewHandler(db, nil)
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
