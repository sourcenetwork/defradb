// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build javaclient

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/node"
	javaclient "github.com/sourcenetwork/defradb/tests/clients/java"
)

// newTestWrapper starts a minimal in-memory node with a single collection and wraps it,
// registering cleanup for both. Used by tests that need a real *javaclient.Wrapper but don't want
// to go through the shared tests/action framework (which has no way to cancel a single
// subscription's context independently of the whole test's lifecycle).
func newTestWrapper(t *testing.T) (*javaclient.Wrapper, context.Context) {
	t.Helper()
	ctx := context.Background()

	n, err := node.New(ctx,
		options.Node().
			SetDisableAPI(true).
			SetDisableP2P(true).
			Store().SetType(options.NodeMemoryStore).
			Node(),
	)
	require.NoError(t, err)
	require.NoError(t, n.Start(ctx))
	t.Cleanup(func() { _ = n.Close(ctx) })

	w, err := javaclient.NewWrapper(n)
	require.NoError(t, err)
	t.Cleanup(w.Close)

	_, err = w.AddCollection(ctx, `type Users { name: String }`)
	require.NoError(t, err)

	return w, ctx
}
