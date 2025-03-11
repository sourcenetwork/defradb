// Copyright 2024 Democratized Data Foundation
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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureContextTxnExplicit(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)

	// set an explicit transaction
	ctx = SetContextTxn(ctx, txn)

	ctx, txn, err = ensureContextTxn(ctx, db, true)
	require.NoError(t, err)

	_, ok := txn.(*explicitTxn)
	assert.True(t, ok)

	_, ok = ctx.Value(txnContextKey{}).(*explicitTxn)
	assert.True(t, ok)
}

func TestEnsureContextTxnImplicit(t *testing.T) {
	ctx := context.Background()

	db, err := newBadgerDB(ctx)
	require.NoError(t, err)

	ctx, txn, err := ensureContextTxn(ctx, db, true)
	require.NoError(t, err)

	_, ok := txn.(*explicitTxn)
	assert.False(t, ok)

	_, ok = ctx.Value(txnContextKey{}).(*explicitTxn)
	assert.False(t, ok)
}
