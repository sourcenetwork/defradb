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
	"github.com/sourcenetwork/immutable"
)

func TestDocumentDeltasDoNotEncodeDocID(t *testing.T) {
	ctx := context.Background()
	txn := datastore.NewTxnFrom(memory.NewDatastore(ctx), lock.NewLockSet(), 1, false, immutable.None[int]())
	ctx = datastore.CtxSetTxn(ctx, txn)

	lww := NewLWW()
	lwwDelta, err := lww.Delta(
		ctx,
		"collection-version",
		NewDocField("name", client.NewFieldValue(client.LWW_REGISTER, client.NewNormalString("Alice"))),
		true,
		1,
	)
	require.NoError(t, err)
	require.NotContains(t, string(lwwDelta.(*LWWDelta).IPLDSchemaBytes()), "docID") //nolint:forcetypeassert

	counter := NewCounter(false)
	counterDelta, err := counter.Delta(
		ctx,
		"collection-version",
		NewDocField("age", client.NewFieldValue(client.P_COUNTER, client.NewNormalInt(1))),
		false,
		1,
	)
	require.NoError(t, err)
	require.NotContains(t, string(counterDelta.(*CounterDelta).IPLDSchemaBytes()), "docID") //nolint:forcetypeassert

	composite := NewDocComposite()
	require.NotContains(t, string(composite.Delta("collection-version", 1).IPLDSchemaBytes()), "docID")
}
