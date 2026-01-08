// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package lock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const timeout = time.Millisecond

func TestLockSet_MultipleLocksForSameTxnKey_DoNotDeadlock(t *testing.T) {
	txn := newTxn(1)
	lockSet := newLockSet[int]()

	lockSet.Lock(txn, 1)
	// The second Lock call must not block - we want identical keys for the same
	// txn to be able to make as many calls as it needs whilst sharing the same
	// underlying lock.
	lockSet.Lock(txn, 1)
}

func TestLockSet_RLockWLockForSameTxnKey_DoNotDeadlock(t *testing.T) {
	txn := newTxn(1)
	lockSet := newLockSet[int]()

	lockSet.RLock(txn, 1)
	// The second Lock call must not block - we want identical keys for the same
	// txn to be able to make as many calls as it needs whilst sharing the same
	// underlying lock.
	lockSet.Lock(txn, 1)
}

func TestLockSet_MultipleWLocksForDifferentTxnSameKey_Deadlocks(t *testing.T) {
	txn1 := newTxn(1)
	txn2 := newTxn(2)
	lockSet := newLockSet[int]()

	lockSet.Lock(txn1, 1)
	require.Never(
		t,
		func() bool {
			// This call should never complete, because txn1 holds the lock
			lockSet.Lock(txn2, 1)
			return true
		},
		timeout,
		timeout,
	)
}

func TestLockSet_MultipleWLocksForDifferentTxnDifferentKeys_DoNotDeadlock(t *testing.T) {
	txn1 := newTxn(1)
	txn2 := newTxn(2)
	lockSet := newLockSet[int]()

	lockSet.Lock(txn1, 1)
	// This call should not be affected by the first lock, as it is for a different key
	lockSet.Lock(txn2, 2)
}

func TestLockSet_RLockWLockForDifferentTxnSameKey_Deadlocks(t *testing.T) {
	txn1 := newTxn(1)
	txn2 := newTxn(2)
	lockSet := newLockSet[int]()

	lockSet.RLock(txn1, 1)
	require.Never(
		t,
		func() bool {
			// This call should never complete, because txn1 holds the lock
			lockSet.Lock(txn2, 1)
			return true
		},
		timeout,
		timeout,
	)
}

func TestLockSet_WLockRLockForDifferentTxnSameKey_Deadlocks(t *testing.T) {
	txn1 := newTxn(1)
	txn2 := newTxn(2)
	lockSet := newLockSet[int]()

	lockSet.Lock(txn1, 1)
	require.Never(
		t,
		func() bool {
			// This call should never complete, because txn1 holds the lock
			lockSet.RLock(txn2, 1)
			return true
		},
		timeout,
		timeout,
	)
}

func TestLockSet_PromotedLockWLockForDifferentTxnSameKey_Deadlocks(t *testing.T) {
	txn1 := newTxn(1)
	txn2 := newTxn(2)
	lockSet := newLockSet[int]()

	lockSet.RLock(txn1, 1)
	// Promote the RLock to a WLock for txn1
	lockSet.Lock(txn1, 1)
	require.Never(
		t,
		func() bool {
			// This call should never complete, because txn1 holds the lock
			lockSet.Lock(txn2, 1)
			return true
		},
		timeout,
		timeout,
	)
}

func TestLockSet_PromotedLockRLockForDifferentTxnSameKey_Deadlocks(t *testing.T) {
	txn1 := newTxn(1)
	txn2 := newTxn(2)
	lockSet := newLockSet[int]()

	lockSet.RLock(txn1, 1)
	// Promote the RLock to a WLock for txn1
	lockSet.Lock(txn1, 1)
	require.Never(
		t,
		func() bool {
			// This call should never complete, because txn1 holds the lock
			lockSet.RLock(txn2, 1)
			return true
		},
		timeout,
		timeout,
	)
}

func TestLockSet_RLockPromotedLockForDifferentTxnSameKey_Deadlocks(t *testing.T) {
	txn1 := newTxn(1)
	txn2 := newTxn(2)
	lockSet := newLockSet[int]()

	lockSet.RLock(txn1, 1)
	lockSet.RLock(txn2, 1)
	require.Never(
		t,
		func() bool {
			// Promote the RLock to a WLock for txn2 - this call should never complete,
			// because txn1 holds the lock
			lockSet.Lock(txn2, 1)
			return true
		},
		timeout,
		timeout,
	)
}

func TestLockSet_DemotedLockRLockForDifferentTxnSameKey_Deadlocks(t *testing.T) {
	txn1 := newTxn(1)
	txn2 := newTxn(2)
	lockSet := newLockSet[int]()

	lockSet.Lock(txn1, 1)
	// 'Demoting' the write lock should have no impact, the lockset does not demote write
	// locks into read locks.  This test is here to enforce that.
	lockSet.RLock(txn1, 1)

	require.Never(
		t,
		func() bool {
			// This call should never complete, because txn1 holds the write lock
			lockSet.RLock(txn2, 1)
			return true
		},
		timeout,
		timeout,
	)
}

func TestLockSet_ClosingTxn_ClearsLocks(t *testing.T) {
	txn1 := newTxn(1)
	txn2 := newTxn(2)
	lockSet := newLockSet[int]()

	lockSet.Lock(txn1, 1)
	txn1.Close()

	// The lockset should have cleared all of txn1's locks on close, unblocking this call
	lockSet.Lock(txn2, 1)
}

func TestLockSet_ClosingTxn_ClearsAllLocks(t *testing.T) {
	txn1 := newTxn(1)
	txn2 := newTxn(2)
	lockSet := newLockSet[int]()

	lockSet.Lock(txn1, 1)
	lockSet.Lock(txn1, 1)
	txn1.Close()

	// The lockset should have cleared all of txn1's locks on close, unblocking this call
	lockSet.Lock(txn2, 1)
}

func TestLockSet_ClosingPromotedTxnLock_ClearsLock(t *testing.T) {
	txn1 := newTxn(1)
	txn2 := newTxn(2)
	lockSet := newLockSet[int]()

	lockSet.RLock(txn1, 1)
	// Promote the RLock to a WLock for txn1
	lockSet.Lock(txn1, 1)
	txn1.Close()

	// The lockset should have cleared all of txn1's locks on close, unblocking this call
	lockSet.Lock(txn2, 1)
}

func TestLockSet_ClosingTxnMultipleTimes_Succedes(t *testing.T) {
	txn1 := newTxn(1)
	lockSet := newLockSet[int]()

	lockSet.Lock(txn1, 1)
	txn1.Close()
	// Discard can be called multiple times, including after txn commit, so it is important
	// to test for this.
	txn1.Close()
}

// todo - This (and the reverse, lock then RLockAll) is the only known case when a
// transaction can deadlock on itself.
//
// It is not known to happen in production at the moment, but it is possible to introduce it
// if not paying attention.
//
// https://github.com/sourcenetwork/defradb/issues/4311
func TestLockSet_RLockAllAndLockSameTxn_Deadlocks(t *testing.T) {
	txn1 := newTxn(1)
	lockSet := newLockSet[int]()

	lockSet.RLockAll(txn1)

	require.Never(
		t,
		func() bool {
			// This call should never complete, because txn1 has read locked everything
			lockSet.Lock(txn1, 1)
			return true
		},
		timeout,
		timeout,
	)
}

func TestLockSet_RLockAllAndLockDifferentTxn_Deadlocks(t *testing.T) {
	txn1 := newTxn(1)
	txn2 := newTxn(2)
	lockSet := newLockSet[int]()

	lockSet.RLockAll(txn1)

	require.Never(
		t,
		func() bool {
			// This call should never complete, because txn1 has read locked everything
			lockSet.Lock(txn2, 1)
			return true
		},
		timeout,
		timeout,
	)
}

type dummyTxn struct {
	id        uint64
	onSuccess []func()
	onError   []func()
	onDiscard []func()
}

var _ = (*dummyTxn)(nil)

func newTxn(id uint64) *dummyTxn {
	return &dummyTxn{
		id: id,
	}
}

// Close is a simple test function that represents either
// a Discard or Commit call both succeding and erroring.
//
// It is permitted for users to discard multiple times,
// including after commit, so multiple calls to this should
// be tested.
func (t *dummyTxn) Close() {
	for _, fn := range t.onSuccess {
		fn()
	}
	for _, fn := range t.onError {
		fn()
	}
	for _, fn := range t.onDiscard {
		fn()
	}
}

func (t *dummyTxn) ID() uint64 {
	return t.id
}

func (t *dummyTxn) OnSuccess(fn func()) {
	t.onSuccess = append(t.onSuccess, fn)
}

func (t *dummyTxn) OnError(fn func()) {
	t.onError = append(t.onError, fn)
}

func (t *dummyTxn) OnDiscard(fn func()) {
	t.onDiscard = append(t.onDiscard, fn)
}
