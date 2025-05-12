// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build js

package node

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNodeStartJS(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	node, err := New(ctx, WithDisableP2P(true), WithDisableAPI(true), WithBadgerInMemory(true))
	require.NoError(t, err)

	err = node.Start(ctx)
	require.NoError(t, err)

	err = node.Close(ctx)
	require.NoError(t, err)
}
