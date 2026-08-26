// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package http

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTruncateCollectionRequestUnmarshalJSON(t *testing.T) {
	var request TruncateCollectionRequest
	require.NoError(t, json.Unmarshal(
		[]byte(`{"filter":{"age":{"_eq":9007199254740993}},"pruneHistory":true}`),
		&request,
	))
	require.Equal(t, map[string]any{
		"age": map[string]any{"_eq": int64(9007199254740993)},
	}, request.Filter)
	require.True(t, request.PruneHistory)

	require.NoError(t, json.Unmarshal([]byte(`{}`), &request))
	require.Nil(t, request.Filter)
	require.False(t, request.PruneHistory)

	require.NoError(t, json.Unmarshal([]byte(`{"filter":null}`), &request))
	require.Nil(t, request.Filter)
}
