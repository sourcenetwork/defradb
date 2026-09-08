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

package external

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/integration/version"
	"github.com/sourcenetwork/defradb/tests/state"
)

// compile-time assertion that Wrapper satisfies the shared testing client
// interface, so the rest of the harness can drive it unchanged.
var _ clients.Client = (*Wrapper)(nil)

// TestExternalWrapper execs the real v1.0.0 release binary and drives it
// end-to-end over HTTP through the Wrapper, proving the wrapper works
// against a real external node.
func TestExternalWrapper(t *testing.T) {
	// Generous budget: a cold cache downloads the release binary, and
	// NewWrapper may retry start/health a few times before succeeding.
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	path, skip, err := version.BinaryPath(ctx, "v1.0.0")
	require.NoError(t, err)
	if skip {
		t.Skip("no v1.0.0 asset for this platform")
	}

	w, err := NewWrapper(ctx, t, path, nil)
	require.NoError(t, err)
	defer w.Close()

	_, err = w.AddCollection(ctx, "type User { name: String }")
	require.NoError(t, err)

	col, err := w.GetCollectionByName(ctx, "User")
	require.NoError(t, err)

	doc, err := client.NewDocFromJSON(ctx, []byte(`{"name":"Alice"}`), col.Version())
	require.NoError(t, err)
	err = col.AddDocument(ctx, doc)
	require.NoError(t, err)

	result := w.ExecRequest(ctx, `query { User { name } }`)
	require.Empty(t, result.GQL.Errors)
	require.Equal(t, map[string]any{
		"User": []any{map[string]any{"name": "Alice"}},
	}, result.GQL.Data)

	addrs, err := w.PeerInfo(ctx, options.PeerInfo())
	require.NoError(t, err)
	require.NotEmpty(t, addrs)

	bus := w.Events()
	require.NotNil(t, bus)
	eventState, err := state.NewEventState(bus)
	require.NoError(t, err)
	require.NotNil(t, eventState)
}
