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

package java

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
)

// TestWrapperClose_ThenNewTxn_ReturnsClosedError guards against calls made after Close using
// nodeObj/handle once the JNI global ref has already been deleted. NewTxn (and every other
// method, via callStore/callGuarded) must reject with errWrapperClosed.
func TestWrapperClose_ThenNewTxn_ReturnsClosedError(t *testing.T) {
	w, ctx := newTestWrapper(t)

	w.Close()

	_, err := w.NewTxn(false)
	require.ErrorContains(t, err, errWrapperClosed)

	_, err = w.GetCollections(ctx, options.GetCollections())
	require.ErrorContains(t, err, errWrapperClosed)
}

// TestWrapperClose_ConcurrentWithNewTxn_NoRace guards against Close deleting nodeObj's JNI global
// ref while NewTxn is concurrently using it. Run with -race. The meaningful assertion isn't which
// of the two wins, it's that NewTxn either succeeds cleanly, or fails with exactly errWrapperClosed.
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
		require.ErrorContains(t, txnErr, errWrapperClosed)
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
	require.ErrorContains(t, err, errWrapperClosed)
}