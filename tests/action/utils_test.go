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

package action

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
)

type retryTestClient struct {
	client.TxnStore
	maxRetries int
}

func (*retryTestClient) Close() {}

func (c *retryTestClient) MaxTxnRetries() int {
	return c.maxRetries
}

func (*retryTestClient) Events() event.Bus {
	return nil
}

func TestWithRetryOnNodeRetriesConflicts(t *testing.T) {
	attempts := 0
	err := withRetryOnNode(&retryTestClient{maxRetries: 3}, func() error {
		attempts++
		if attempts < 3 {
			return corekv.ErrTxnConflict
		}
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 3, attempts)
}

func TestWithRetryOnNodeReturnsLastConflict(t *testing.T) {
	attempts := 0
	err := withRetryOnNode(&retryTestClient{maxRetries: 2}, func() error {
		attempts++
		return corekv.ErrTxnConflict
	})

	require.ErrorIs(t, err, corekv.ErrTxnConflict)
	require.Equal(t, 2, attempts)
}

func TestWithRetryOnNodeAttemptsOnceWithNoRetries(t *testing.T) {
	attempts := 0
	err := withRetryOnNode(&retryTestClient{}, func() error {
		attempts++
		return nil
	})

	require.NoError(t, err)
	require.Equal(t, 1, attempts)
}
