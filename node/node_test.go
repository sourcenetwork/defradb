// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build !js

package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client/options"
)

func TestPurgeAndRestartWithDevModeDisabled(t *testing.T) {
	ctx := context.Background()

	n, err := New(ctx,
		options.Node().
			SetDisableAPI(true).
			SetDisableP2P(true).
			Store().SetPath(t.TempDir()).
			Node(),
	)
	require.NoError(t, err)

	err = n.Start(ctx)
	require.NoError(t, err)

	err = n.PurgeAndRestart(ctx)
	require.ErrorIs(t, err, ErrPurgeWithDevModeDisabled)
}

func TestPurgeAndRestartWithDevModeEnabled(t *testing.T) {
	ctx := context.Background()

	n, err := New(ctx,
		options.Node().
			SetDisableAPI(true).
			SetDisableP2P(true).
			SetEnableDevelopment(true).
			Store().SetPath(t.TempDir()).
			Node(),
	)
	require.NoError(t, err)

	err = n.Start(ctx)
	require.NoError(t, err)

	_, err = n.DB.AddSchema(ctx, "type User { name: String }")
	require.NoError(t, err)

	err = n.PurgeAndRestart(ctx)
	require.NoError(t, err)

	schemas, err := n.DB.GetCollections(ctx)
	require.NoError(t, err)

	assert.Len(t, schemas, 0)
}
