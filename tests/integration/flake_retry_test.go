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

package tests

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlakeRetry_WithZeroRetries_ShouldRunOnce(t *testing.T) {
	var runCount atomic.Int32
	passed := runTestWithRetry(t, 0, func(st testing.TB) {
		runCount.Add(1)
	})
	assert.True(t, passed)
	assert.Equal(t, int32(1), runCount.Load())
}

func TestFlakeRetry_WithRetriesAndPassingTest_ShouldRunOnce(t *testing.T) {
	var runCount atomic.Int32
	passed := runTestWithRetry(t, 3, func(st testing.TB) {
		runCount.Add(1)
	})
	assert.True(t, passed)
	assert.Equal(t, int32(1), runCount.Load())
}

func TestFlakeRetry_WithRetriesAndFlakyTest_ShouldRetryUntilPass(t *testing.T) {
	var runCount atomic.Int32
	passed := runTestWithRetry(t, 3, func(st testing.TB) {
		count := runCount.Add(1)
		if count < 3 {
			st.Fatal("flaky failure")
		}
	})
	assert.True(t, passed)
	assert.Equal(t, int32(3), runCount.Load())
}

func TestFlakeRetry_WithRetriesAndAlwaysFailingTest_ShouldExhaustRetries(t *testing.T) {
	var runCount atomic.Int32
	passed := runTestWithRetry(t, 2, func(st testing.TB) {
		runCount.Add(1)
		st.Fatal("always fails")
	})
	assert.False(t, passed)
	// 1 initial + 2 retries = 3 total
	assert.Equal(t, int32(3), runCount.Load())
}

func TestFlakeRetry_WithErrorfFailure_ShouldRetryUntilPass(t *testing.T) {
	var runCount atomic.Int32
	passed := runTestWithRetry(t, 3, func(st testing.TB) {
		count := runCount.Add(1)
		if count < 2 {
			st.Errorf("non-fatal flaky failure")
		}
	})
	assert.True(t, passed)
	assert.Equal(t, int32(2), runCount.Load())
}
