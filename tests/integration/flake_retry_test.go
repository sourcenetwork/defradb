// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
