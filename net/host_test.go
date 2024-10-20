// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package net

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetupHostWithDefaultOptions(t *testing.T) {
	h, dht, err := setupHost(context.Background(), DefaultOptions())
	require.NoError(t, err)

	require.NotNil(t, h)
	require.NotNil(t, dht)

	err = h.Close()
	require.NoError(t, err)
}
