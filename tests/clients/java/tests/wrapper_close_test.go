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

//go:build javaclient

package tests

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	javaclient "github.com/sourcenetwork/defradb/tests/clients/java"
)

// TestWrapperClose_ThenNewTxn_ReturnsClosedError guards against calls made after Close using
// nodeObj/handle once the JNI global ref has already been deleted. NewTxn (and every other
// method, via callStore/callGuarded) must reject with ErrWrapperClosed.
func TestWrapperClose_ThenNewTxn_ReturnsClosedError(t *testing.T) {
	w, ctx := newTestWrapper(t)

	w.Close()

	_, err := w.NewTxn(false)
	require.ErrorContains(t, err, javaclient.ErrWrapperClosed)

	_, err = w.GetCollections(ctx, options.GetCollections())
	require.ErrorContains(t, err, javaclient.ErrWrapperClosed)
}

// TestWrapperClose_ConcurrentWithNewTxn_NoRace guards against Close deleting nodeObj's JNI global
// ref while NewTxn is concurrently using it. Run with -race. The meaningful assertion isn't which
// of the two wins, it's that NewTxn either succeeds cleanly, or fails with exactly ErrWrapperClosed.
// There should never be a raw JNI/native error from touching a stale reference, which is what an
// unsynchronized race would produce instead.
func TestWrapperClose_ConcurrentWithNewTxn_NoRace(t *testing.T) {
	w, _ := newTestWrapper(t)

	var wg sync.WaitGroup
	var txn client.Txn
	var txnErr error
	wg.Add(2)
	go func() {
		defer wg.Done()
		w.Close()
	}()
	go func() {
		defer wg.Done()
		txn, txnErr = w.NewTxn(false)
	}()
	wg.Wait()

	if txnErr == nil {
		txn.Discard()
	} else {
		require.ErrorContains(t, txnErr, javaclient.ErrWrapperClosed)
	}
}

// TestWrapperClose_ConcurrentDoubleClose_NoRace guards against two overlapping Close calls both
// trying to finalize (and delete the JNI global ref for) the same nodeObj. Run with -race. nodeMu
// must serialize them so only one actually runs the close sequence, leaving the wrapper in a
// single, consistent closed state afterwards rather than double-deleting the reference.
func TestWrapperClose_ConcurrentDoubleClose_NoRace(t *testing.T) {
	w, _ := newTestWrapper(t)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		w.Close()
	}()
	go func() {
		defer wg.Done()
		w.Close()
	}()
	wg.Wait()

	_, err := w.NewTxn(false)
	require.ErrorContains(t, err, javaclient.ErrWrapperClosed)
}

// TestWrapperClose_WhileSubscribed_NoRace guards against wrapSubscriptionAsChannel's
// subscription-polling goroutine racing with Close deleting w.nodeObj's JNI global ref. The
// subscription's context is never cancelled here, so the poll loop is still (or about to be)
// calling PollSubscriptionNative/CloseSubscriptionNative on w.nodeObj when t.Cleanup's Close runs
// at the end of this test - Close must not be able to delete the ref out from under an in-flight
// or about-to-start poll. Run with -race. There is nothing to assert on beyond "this doesn't crash
// or race": the subscription's own contents aren't the point, so no attempt is made to read from
// (or cancel) it before the test returns.
func TestWrapperClose_WhileSubscribed_NoRace(t *testing.T) {
	w, ctx := newTestWrapper(t)

	result := w.ExecRequest(ctx, `subscription { Users { name } }`, options.ExecRequest())
	require.Empty(t, result.GQL.Errors)
	require.NotNil(t, result.Subscription)
}
