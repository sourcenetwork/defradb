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

package cli

import (
	"context"
	"net/http/httptest"
	"testing"

	badgerds "github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/badger"
	"github.com/sourcenetwork/corelog"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	httpapi "github.com/sourcenetwork/defradb/http"
	"github.com/sourcenetwork/defradb/internal/db"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
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
	adminInfo, err := acpDB.NewNACInfo(ctx, "", false)
	if err != nil {
		return nil, errors.Wrap("failed to setup node access control info", err)
	}
	db, err := db.NewDB(ctx, rootstore, adminInfo)
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
