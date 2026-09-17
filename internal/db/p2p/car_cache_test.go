// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package p2p

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// countingBuild returns a build that yields data and counts its calls.
func countingBuild(data []byte, err error, calls *int) func() ([]byte, error) {
	return func() ([]byte, error) {
		*calls++
		return data, err
	}
}

func TestCARCache_BuildsOnceThenServesFromCache(t *testing.T) {
	c := newCARCache(1 << 10)
	calls := 0

	data, shared, err := c.getOrBuild(context.Background(), "a", countingBuild([]byte("car"), nil, &calls))
	require.NoError(t, err)
	require.False(t, shared)
	require.Equal(t, []byte("car"), data)

	data, shared, err = c.getOrBuild(context.Background(), "a", countingBuild([]byte("other"), nil, &calls))
	require.NoError(t, err)
	require.True(t, shared)
	require.Equal(t, []byte("car"), data)
	require.Equal(t, 1, calls)
}

// A request arriving while the head is being built takes that build's result rather than
// starting its own.
func TestCARCache_ConcurrentRequestSharesTheBuild(t *testing.T) {
	c := newCARCache(1 << 10)
	started := make(chan struct{})
	release := make(chan struct{})

	go func() {
		_, _, _ = c.getOrBuild(context.Background(), "a", func() ([]byte, error) {
			close(started)
			<-release
			return []byte("car"), nil
		})
	}()
	<-started

	result := make(chan []byte)
	go func() {
		data, shared, err := c.getOrBuild(context.Background(), "a", func() ([]byte, error) {
			return nil, errors.New("second build started")
		})
		require.NoError(t, err)
		require.True(t, shared)
		result <- data
	}()

	close(release)
	require.Equal(t, []byte("car"), <-result)
}

// A build that panics hands the requests waiting on it an error rather than an empty CAR they
// would report as served, lets the panic carry on in the goroutine that built, and caches nothing,
// so the next request builds afresh.
func TestCARCache_BuildThatPanicsFailsItsWaiters(t *testing.T) {
	c := newCARCache(1 << 10)
	started := make(chan struct{})
	release := make(chan struct{})
	recovered := make(chan any, 1)

	go func() {
		defer func() { recovered <- recover() }()
		_, _, _ = c.getOrBuild(context.Background(), "a", func() ([]byte, error) {
			close(started)
			<-release
			panic("build failed")
		})
	}()
	<-started

	// Waiters return exactly this build's data and error once done is closed.
	c.mu.Lock()
	b := c.building["a"]
	c.mu.Unlock()
	require.NotNil(t, b)

	close(release)
	<-b.done

	require.Equal(t, "build failed", <-recovered)
	require.Nil(t, b.data)
	require.ErrorIs(t, b.err, errCARBuildDidNotReturn)
	require.Empty(t, c.entries)
	require.Empty(t, c.building)

	calls := 0
	data, shared, err := c.getOrBuild(context.Background(), "a", countingBuild([]byte("car"), nil, &calls))
	require.NoError(t, err)
	require.False(t, shared)
	require.Equal(t, []byte("car"), data)
	require.Equal(t, 1, calls)
}

// A waiter whose context ends stops waiting without disturbing the build it was waiting on.
func TestCARCache_WaiterGivesUpOnItsContext(t *testing.T) {
	c := newCARCache(1 << 10)
	started := make(chan struct{})
	release := make(chan struct{})
	built := make(chan struct{})

	go func() {
		_, _, _ = c.getOrBuild(context.Background(), "a", func() ([]byte, error) {
			close(started)
			<-release
			return []byte("car"), nil
		})
		close(built)
	}()
	<-started

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := c.getOrBuild(ctx, "a", func() ([]byte, error) { return nil, nil })
	require.ErrorIs(t, err, context.Canceled)

	close(release)
	<-built
	calls := 0
	data, shared, err := c.getOrBuild(context.Background(), "a", countingBuild(nil, nil, &calls))
	require.NoError(t, err)
	require.True(t, shared)
	require.Equal(t, []byte("car"), data)
	require.Zero(t, calls)
}

// A failed build is not kept: the block it lacked may arrive, so the next request builds again.
func TestCARCache_DoesNotKeepAFailedBuild(t *testing.T) {
	c := newCARCache(1 << 10)
	calls := 0

	_, _, err := c.getOrBuild(context.Background(), "a", countingBuild(nil, errors.New("missing"), &calls))
	require.Error(t, err)

	data, shared, err := c.getOrBuild(context.Background(), "a", countingBuild([]byte("car"), nil, &calls))
	require.NoError(t, err)
	require.False(t, shared)
	require.Equal(t, []byte("car"), data)
	require.Equal(t, 2, calls)
}

// Past the byte budget the least recently used entry goes, and a use refreshes an entry.
func TestCARCache_EvictsLeastRecentlyUsedPastTheBudget(t *testing.T) {
	c := newCARCache(10)
	calls := 0
	ctx := context.Background()

	_, _, _ = c.getOrBuild(ctx, "a", countingBuild([]byte("aaaa"), nil, &calls))
	_, _, _ = c.getOrBuild(ctx, "b", countingBuild([]byte("bbbb"), nil, &calls))
	_, _, _ = c.getOrBuild(ctx, "a", countingBuild(nil, nil, &calls)) // a is now the most recent
	_, _, _ = c.getOrBuild(ctx, "c", countingBuild([]byte("cccc"), nil, &calls))

	require.Contains(t, c.entries, "a")
	require.NotContains(t, c.entries, "b")
	require.Contains(t, c.entries, "c")
	require.Equal(t, 8, c.bytes)
	require.Equal(t, 3, calls)
}

func TestCARCache_DoesNotKeepACAROverTheWholeBudget(t *testing.T) {
	c := newCARCache(4)
	calls := 0

	data, _, err := c.getOrBuild(context.Background(), "a", countingBuild([]byte("too large"), nil, &calls))
	require.NoError(t, err)
	require.Equal(t, []byte("too large"), data)
	require.Empty(t, c.entries)
	require.Zero(t, c.bytes)
}

// Serving the same head to two peers builds its CAR once.
func TestServeCARs_SecondRequestForAHeadIsACacheHit(t *testing.T) {
	child := undecodableBlock(t, "field block")
	_, root := compositeLinking(t, child.Cid())
	p := carFixture(t, root, child)
	serveEveryBlock(t, p)
	p.carCache = newCARCache(carCacheMaxBytes)

	first := p.serveCARs(context.Background(), "peer-a", [][]byte{root.Cid().Bytes()})
	second := p.serveCARs(context.Background(), "peer-b", [][]byte{root.Cid().Bytes()})

	require.Equal(t, first.CARs, second.CARs)
	require.Equal(t, int64(1), p.statCARBuilt.Load())
	require.Equal(t, int64(1), p.statCARCacheHits.Load())
}
