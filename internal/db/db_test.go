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

	badgerds "github.com/dgraph-io/badger/v4"
	"github.com/sourcenetwork/corekv/badger"
	"github.com/stretchr/testify/require"
)

func newBadgerDB(ctx context.Context) (*DB, error) {
	rootstore, err := badger.NewDatastore("", badgerds.DefaultOptions("").WithInMemory(true))
	if err != nil {
		return nil, err
	}
	return newDB(ctx, rootstore, nil)
}

func TestNewDB(t *testing.T) {
	ctx := context.Background()
	rootstore, err := badger.NewDatastore("", badgerds.DefaultOptions("").WithInMemory(true))
	require.NoError(t, err)

	_, err = NewDB(ctx, rootstore, nil)
	require.NoError(t, err)
}
