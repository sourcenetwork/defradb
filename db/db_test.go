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

	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
)

func newMemoryDB(ctx context.Context) (*db, error) {
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		return nil, err
	}
	return newDB(ctx, rootstore)
}

func TestNewDB(t *testing.T) {
	ctx := context.Background()
	opts := badgerds.Options{Options: badger.DefaultOptions("").WithInMemory(true)}
	rootstore, err := badgerds.NewDatastore("", &opts)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = NewDB(ctx, rootstore)
	if err != nil {
		t.Error(err)
	}
}
