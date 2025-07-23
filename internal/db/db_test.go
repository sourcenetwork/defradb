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

	"github.com/sourcenetwork/defradb/acp/dac"
)

func newBadgerDB(ctx context.Context) (*DB, error) {
	rootstore, err := badger.NewDatastore("", badgerds.DefaultOptions("").WithInMemory(true))
	if err != nil {
		return nil, err
	}

	adminInfo, err := NewAdminInfo(ctx, "", false)
	if err != nil {
		return nil, err
	}
	return newDB(ctx, rootstore, adminInfo, dac.NoDocumentACP, nil)
}

func TestNewDB(t *testing.T) {
	ctx := context.Background()
	rootstore, err := badger.NewDatastore("", badgerds.DefaultOptions("").WithInMemory(true))
	require.NoError(t, err)

	adminInfo, err := NewAdminInfo(ctx, "", false)
	require.NoError(t, err)

	_, err = NewDB(ctx, rootstore, adminInfo, dac.NoDocumentACP, nil)
	require.NoError(t, err)
}
