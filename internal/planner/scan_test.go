// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/db/lock"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
	"github.com/sourcenetwork/immutable"
)

func TestDocIDFilterAliasExpansion(t *testing.T) {
	ctx := context.Background()
	txn := datastore.NewTxnFrom(memory.NewDatastore(ctx), lock.NewLockSet(), 1, false, immutable.None[int]())
	defer txn.Discard()
	ctx = datastore.CtxSetTxn(ctx, txn)

	const (
		collectionShortID = uint32(1)
		shortDocID        = uint32(7)
		publicDocID       = "bae-public-doc"
		legacyDocID       = "bae-legacy-doc"
	)
	require.NoError(t, id.SetDocIDAlias(ctx, collectionShortID, shortDocID, publicDocID))
	require.NoError(t, id.SetDocIDAlias(ctx, collectionShortID, shortDocID, legacyDocID))

	values, err := docIDFilterValues(ctx, publicDocID)
	require.NoError(t, err)
	require.ElementsMatch(t, []any{publicDocID, legacyDocID}, values)

	expandedValues, err := expandDocIDAliasValues(ctx, []any{publicDocID, legacyDocID, 123})
	require.NoError(t, err)
	require.ElementsMatch(t, []any{publicDocID, legacyDocID, 123}, expandedValues)

	opMap := map[connor.FilterKey]any{
		mapper.FilterEqOp: publicDocID,
	}
	expandedOpMap, changed, err := expandDocIDAliasesInOpMap(ctx, opMap)
	require.NoError(t, err)
	require.True(t, changed)
	require.NotContains(t, expandedOpMap, mapper.FilterEqOp)
	require.ElementsMatch(t, []any{publicDocID, legacyDocID}, expandedOpMap[mapper.FilterInOp])

	values, err = docIDFilterValues(ctx, "bae-unknown-doc")
	require.NoError(t, err)
	require.Equal(t, []any{"bae-unknown-doc"}, values)
}
