// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package node

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithDisableP2P(t *testing.T) {
	options := &Options{}
	WithDisableP2P(true)(options)
	assert.Equal(t, true, options.disableP2P)
}

func TestWithDisableAPI(t *testing.T) {
	options := &Options{}
	WithDisableAPI(true)(options)
	assert.Equal(t, true, options.disableAPI)
}

func TestWithEnableDevelopment(t *testing.T) {
	options := &Options{}
	WithEnableDevelopment(true)(options)
	assert.Equal(t, true, options.enableDevelopment)
}

func TestPurgeAndRestartWithDevModeDisabled(t *testing.T) {
	ctx := context.Background()

	opts := []Option{
		WithDisableAPI(true),
		WithDisableP2P(true),
		WithStorePath(t.TempDir()),
	}

	n, err := New(ctx, opts...)
	require.NoError(t, err)

	err = n.Start(ctx)
	require.NoError(t, err)

	err = n.PurgeAndRestart(ctx)
	require.ErrorIs(t, err, ErrPurgeWithDevModeDisabled)
}

func TestPurgeAndRestartWithDevModeEnabled(t *testing.T) {
	ctx := context.Background()

	opts := []Option{
		WithDisableAPI(true),
		WithDisableP2P(true),
		WithStorePath(t.TempDir()),
		WithEnableDevelopment(true),
	}

	n, err := New(ctx, opts...)
	require.NoError(t, err)

	err = n.Start(ctx)
	require.NoError(t, err)

	_, err = n.DB.AddSchema(ctx, "type User { name: String }")
	require.NoError(t, err)

	err = n.PurgeAndRestart(ctx)
	require.NoError(t, err)

	schemas, err := n.DB.GetSchemas(ctx, client.SchemaFetchOptions{})
	require.NoError(t, err)

	assert.Len(t, schemas, 0)
}
