// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTxnFrom(t *testing.T) {
	ctx := context.Background()
	rootstore := getBadgerTxnDB(t)

	txn, err := NewTxnFrom(ctx, rootstore, 0, false)
	require.NoError(t, err)

	err = txn.Commit(ctx)
	require.NoError(t, err)
}

func TestOnSuccess(t *testing.T) {
	ctx := context.Background()
	rootstore := getBadgerTxnDB(t)

	txn, err := NewTxnFrom(ctx, rootstore, 0, false)
	require.NoError(t, err)

	txn.OnSuccess(nil)

	text := "Source"
	txn.OnSuccess(func() {
		text += " Inc"
	})
	err = txn.Commit(ctx)
	require.NoError(t, err)

	require.Equal(t, text, "Source Inc")
}

func TestOnError(t *testing.T) {
	ctx := context.Background()
	rootstore := getBadgerTxnDB(t)

	txn, err := NewTxnFrom(ctx, rootstore, 0, false)
	require.NoError(t, err)

	txn.OnError(nil)

	text := "Source"
	txn.OnError(func() {
		text += " Inc"
	})

	rootstore.Close()
	require.NoError(t, err)

	err = txn.Commit(ctx)
	require.Error(t, err)

	require.Equal(t, text, "Source Inc")
}
