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

	"github.com/sourcenetwork/defradb/db/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetContextImplicitTxn(t *testing.T) {
	ctx := context.Background()

	db, err := newMemoryDB(ctx)
	require.NoError(t, err)

	txn, err := getContextTxn(ctx, db, true)
	require.NoError(t, err)

	// txn should be implicit
	_, ok := txn.(*explicitTxn)
	assert.False(t, ok)
}

func TestGetContextExplicitTxn(t *testing.T) {
	ctx := context.Background()

	db, err := newMemoryDB(ctx)
	require.NoError(t, err)

	txn, err := db.NewTxn(ctx, true)
	require.NoError(t, err)

	// create a session with a transaction
	sess := session.New(ctx).WithTxn(txn)

	out, err := getContextTxn(sess, db, true)
	require.NoError(t, err)

	// txn should be explicit
	_, ok := out.(*explicitTxn)
	assert.True(t, ok)
}
