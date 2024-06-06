// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"testing"

	badger "github.com/sourcenetwork/badger/v4"

	"github.com/sourcenetwork/defradb/acp"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/datastore/memory"
)

func newMemoryDB(ctx context.Context) (*db, error) {
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		return nil, err
	}
	return newDB(ctx, rootstore, acp.NoACP, nil)
}

func newDefraMemoryDB(ctx context.Context) (*db, error) {
	rootstore := memory.NewDatastore(ctx)
	return newDB(ctx, rootstore, acp.NoACP)
}

func TestNewDB(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = NewDB(ctx, rootstore, acp.NoACP, nil)
	if err != nil {
		t.Error(err)
	}
}
