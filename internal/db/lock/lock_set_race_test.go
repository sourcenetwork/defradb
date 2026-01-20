// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// The tests in this file contain deliberate race conditions, and so must not execute when
// running tests with Golang's race detector.

//go:build !race

package lock

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// WARNING - If this test detects a bug, it will become flaky! Never ignore flakiness in this
// test - it means something has broken in the lockset and there could be some fairly unpleasent
// bugs in production.
func TestLockSet_ConcurrentRWLockRLockForDifferentTxnSameKey_Deadlocks(t *testing.T) {
	txn1 := newTxn(1)
	txn2 := newTxn(2)
	lockSet := newLockSet[int]()

	wait := sync.RWMutex{}
	// Lock the wait in order to make sure all concurrent locks start as close as possible to
	// the same time.
	wait.Lock()

	// Create some routines read and write locking txn1.
	go func() {
		wait.RLock()
		lockSet.RLock(txn1, 1)
	}()
	go func() {
		wait.RLock()
		lockSet.Lock(txn1, 1)
	}()
	go func() {
		wait.RLock()
		lockSet.RLock(txn1, 1)
	}()
	go func() {
		wait.RLock()
		lockSet.RLock(txn1, 1)
	}()
	go func() {
		wait.RLock()
		lockSet.Lock(txn1, 1)
	}()
	go func() {
		wait.RLock()
		lockSet.RLock(txn1, 1)
	}()

	// Unlock the wait, allowing the child routines to start read-write locking txn1
	wait.Unlock()

	// Wait a small amount of time to let the RWLocking routines complete
	// without having to bother writing actual synchronizing code.
	time.Sleep(time.Millisecond)

	require.Never(
		t,
		func() bool {
			// This call should never complete, because no matter in what order the child routines
			// executed, and no mater how concurrent they actually were, txn1 *must* hold the write
			// lock to key `1` by the time they complete - causing this call for txn2 to block.
			lockSet.RLock(txn2, 1)
			return true
		},
		timeout,
		timeout,
	)
}

// WARNING - If this test detects a bug, it will become flaky! Never ignore flakiness in this
// test - it means something has broken in the lockset and there could be some fairly unpleasent
// bugs in production.
func TestLockSet_ConcurrentRWLockRLockForSameTxnKey_ClosesCorrectly(t *testing.T) {
	txn1 := newTxn(1)
	txn2 := newTxn(2)
	lockSet := newLockSet[int]()

	wait := sync.RWMutex{}
	// Lock the wait in order to make sure all concurrent locks start as close as possible to
	// the same time.
	wait.Lock()

	// Create some routines read and write locking txn1.
	go func() {
		wait.RLock()
		lockSet.RLock(txn1, 1)
	}()
	go func() {
		wait.RLock()
		lockSet.Lock(txn1, 1)
	}()
	go func() {
		wait.RLock()
		lockSet.RLock(txn1, 1)
	}()
	go func() {
		wait.RLock()
		lockSet.RLock(txn1, 1)
	}()
	go func() {
		wait.RLock()
		lockSet.Lock(txn1, 1)
	}()
	go func() {
		wait.RLock()
		lockSet.RLock(txn1, 1)
	}()

	// Unlock the wait, allowing the child routines to start read-write locking txn1
	wait.Unlock()

	// Wait a small amount of time to let the RWLocking routines complete
	// without having to bother writing actual synchronizing code.
	time.Sleep(time.Millisecond)

	// Ensure that the transaction closes correctly, regardless of the order and concurrency
	// of operations.  This is important to test as [R]Unlock will panic if it is not locked.
	txn1.Close()

	// Ensure that all locks have been unlocked by acquiring a write lock using another txn.
	lockSet.Lock(txn2, 1)
}
