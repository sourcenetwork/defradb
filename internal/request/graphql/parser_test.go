// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package graphql

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv/memory"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/lock"
)

// Parsing is safe while a schema change commits on another goroutine, and the change is used once it commits.
func TestParseWhileSetSchemaCommits(t *testing.T) {
	ctx := context.Background()
	p, err := NewParser(false)
	require.NoError(t, err)

	request, err := p.BuildRequestAST(ctx, "{ User { name } }")
	require.NoError(t, err)

	txn := datastore.NewTxnFrom(memory.NewDatastore(ctx), lock.NewLockSet(), 1, false, immutable.None[int]())
	collections := []client.CollectionVersion{{
		Name:   "User",
		Fields: []client.CollectionFieldDescription{{Name: "name", Kind: client.FieldKind_NILLABLE_STRING}},
	}}
	require.NoError(t, p.SetSchema(datastore.CtxSetTxn(ctx, txn), collections))

	committed := make(chan error)
	go func() {
		committed <- txn.Commit()
	}()
	p.Parse(ctx, request, &client.GQLOptions{})
	require.NoError(t, <-committed)

	_, errs := p.Parse(ctx, request, &client.GQLOptions{})
	require.Empty(t, errs)
}
