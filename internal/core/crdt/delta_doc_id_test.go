// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crdt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/immutable"
)

func TestDocumentDeltasDoNotEncodeDocID(t *testing.T) {
	ctx := context.Background()
	txn := datastore.NewTxnFrom(memory.NewDatastore(ctx), lock.NewLockSet(), 1, false, immutable.None[int]())
	ctx = datastore.CtxSetTxn(ctx, txn)
	store := txn.Datastore()

	key := keys.DataStoreKey{
		CollectionShortID: 1,
		DocShortID:        42,
		FieldID:           "1",
	}

	lww := NewLWW(store, "collection-version", key, "name")
	lwwDelta, err := lww.Delta(
		ctx,
		NewDocField("name", client.NewFieldValue(client.LWW_REGISTER, client.NewNormalString("Alice"))),
	)
	require.NoError(t, err)
	require.NotContains(t, string(lwwDelta.(*LWWDelta).IPLDSchemaBytes()), "docID") //nolint:forcetypeassert

	counter := NewCounter(store, "collection-version", key, "age", false, client.FieldKind_NILLABLE_INT)
	counterDelta, err := counter.Delta(
		ctx,
		NewDocField("age", client.NewFieldValue(client.P_COUNTER, client.NewNormalInt(1))),
	)
	require.NoError(t, err)
	require.NotContains(t, string(counterDelta.(*CounterDelta).IPLDSchemaBytes()), "docID") //nolint:forcetypeassert

	composite := NewDocComposite(store, "collection-version", key)
	require.NotContains(t, string(composite.Delta().IPLDSchemaBytes()), "docID")
}
