// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTruncateFilter(t *testing.T) {
	filter, err := parseTruncateFilter(`{"name":{"_eq":"Alice"}}`)
	require.NoError(t, err)
	require.Equal(t, map[string]any{"name": map[string]any{"_eq": "Alice"}}, filter)

	_, err = parseTruncateFilter("null")
	require.EqualError(t, err, "filter cannot be null")

	_, err = parseTruncateFilter("{")
	require.Error(t, err)
}
