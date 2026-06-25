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
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseTxnTTLFlag(t *testing.T) {
	t.Run("duration", func(t *testing.T) {
		txnTTL, err := parseTxnTTLFlag("150ms")

		require.NoError(t, err)
		require.Equal(t, 150*time.Millisecond, txnTTL)
	})

	t.Run("seconds", func(t *testing.T) {
		txnTTL, err := parseTxnTTLFlag("2")

		require.NoError(t, err)
		require.Equal(t, 2*time.Second, txnTTL)
	})

	t.Run("zero", func(t *testing.T) {
		txnTTL, err := parseTxnTTLFlag("0")

		require.NoError(t, err)
		require.Zero(t, txnTTL)
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := parseTxnTTLFlag("soon")

		require.Error(t, err)
	})
}
