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

package tests

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/node"
	"github.com/sourcenetwork/defradb/tests/action"
)

// Store type selection and the shared node options live in the action package
// alongside node setup. They are aliased here so the harness and existing tests
// keep reading a single source of truth rather than a second copy.
const (
	BadgerIMType   = action.BadgerIMType
	DefraIMType    = action.DefraIMType
	BadgerFileType = action.BadgerFileType
	LevelStoreType = action.LevelStoreType
)

func NewBadgerMemoryDB(ctx context.Context) (node.DB, error) {
	opts := options.Node().
		SetDisableP2P(true).
		SetDisableAPI(true).
		SetEnableDevelopment(true).
		Store().SetBadgerInMemory(true).
		Node()

	n, err := node.New(ctx, opts)
	if err != nil {
		return nil, err
	}
	err = n.Start(ctx)
	if err != nil {
		return nil, err
	}
	return n.DB, err
}

func NewBadgerFileDB(ctx context.Context, t testing.TB) (node.DB, error) {
	path := t.TempDir()

	opts := options.Node().
		SetDisableP2P(true).
		SetDisableAPI(true).
		SetEnableDevelopment(true).
		Store().SetPath(path).
		Node()

	n, err := node.New(ctx, opts)
	if err != nil {
		return nil, err
	}
	err = n.Start(ctx)
	if err != nil {
		return nil, err
	}
	return n.DB, err
}
