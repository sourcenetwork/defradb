// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package id

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/internal/keys"
)

func TestGetShortCollectionIDFromStoreFallsBackWithoutCache(t *testing.T) {
	ctx := context.Background()
	txn := newDocumentIDTestTxn(ctx)
	defer txn.Discard()

	const (
		collectionID      = "bae-collection"
		collectionShortID = "12"
	)
	err := txn.Systemstore().Set(ctx, keys.NewCollectionID(collectionID).Bytes(), []byte(collectionShortID))
	require.NoError(t, err)

	got, err := GetShortCollectionIDFromStore(ctx, collectionID, txn.Systemstore())
	require.NoError(t, err)
	require.Equal(t, uint32(12), got)
}
