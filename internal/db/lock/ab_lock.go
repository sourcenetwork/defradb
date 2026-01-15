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
	"sync"
)

// abLock provides a mechanism that allows two competing action sets to block
// the other without blocking other members of its set.
//
// LockA will prevent LockB from being acquired by other transactions, but it will not prevent
// other LockAs from being acquired, and vice versa.
//
// When a transaction already holds a lock of one type and tries to acquire the opposite type,
// it will wait for OTHER transactions' locks to be released, but not its own. This prevents
// self-deadlock while maintaining proper lock semantics - the transaction will hold both
// lock types and properly block other transactions.
//
// Acquired locks are released on transaction close.
type abLock struct {
	mu       sync.Mutex
	initOnce sync.Once

	// cond is used to signal when lock counts change
	cond *sync.Cond

	// aCount tracks the total number of A-locks held (across all transactions)
	aCount int
	// bCount tracks the total number of B-locks held (across all transactions)
	bCount int

	// aHolders tracks how many A-locks each transaction holds (by txn ID)
	aHolders map[uint64]int
	// bHolders tracks how many B-locks each transaction holds (by txn ID)
	bHolders map[uint64]int
}

// init initializes the abLock if needed (lazy initialization)
func (l *abLock) init() {
	l.initOnce.Do(func() {
		l.cond = sync.NewCond(&l.mu)
		l.aHolders = make(map[uint64]int)
		l.bHolders = make(map[uint64]int)
	})
}

// LockA acquires an A-lock for the given transaction.
//
// This will block until all B-locks held by OTHER transactions are released.
// If this transaction already holds B-locks, those are excluded from the wait.
func (l *abLock) LockA(txn txn) {
	l.init()
	l.mu.Lock()

	txnID := txn.ID()

	// Count of B-locks held by OTHER transactions
	otherBCount := l.bCount - l.bHolders[txnID]

	// Wait until all OTHER transactions' B-locks are released
	for otherBCount > 0 {
		l.cond.Wait()
		otherBCount = l.bCount - l.bHolders[txnID]
	}

	// Acquire the A-lock
	l.aCount++
	l.aHolders[txnID]++

	// If this is the first lock of any kind for this transaction,
	// register cleanup callbacks
	needRegister := l.aHolders[txnID] == 1 && l.bHolders[txnID] == 0

	l.mu.Unlock()

	if needRegister {
		l.registerCleanup(txn, txnID)
	}
}

// LockB acquires a B-lock for the given transaction.
//
// This will block until all A-locks held by OTHER transactions are released.
// If this transaction already holds A-locks, those are excluded from the wait.
func (l *abLock) LockB(txn txn) {
	l.init()
	l.mu.Lock()

	txnID := txn.ID()

	// Count of A-locks held by OTHER transactions
	otherACount := l.aCount - l.aHolders[txnID]

	// Wait until all OTHER transactions' A-locks are released
	for otherACount > 0 {
		l.cond.Wait()
		otherACount = l.aCount - l.aHolders[txnID]
	}

	// Acquire the B-lock
	l.bCount++
	l.bHolders[txnID]++

	// If this is the first lock of any kind for this transaction,
	// register cleanup callbacks
	needRegister := l.bHolders[txnID] == 1 && l.aHolders[txnID] == 0

	l.mu.Unlock()

	if needRegister {
		l.registerCleanup(txn, txnID)
	}
}

// registerCleanup registers transaction close callbacks to release all locks
// held by this transaction.
func (l *abLock) registerCleanup(txn txn, txnID uint64) {
	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			l.mu.Lock()
			// Release all A-locks held by this transaction
			if count := l.aHolders[txnID]; count > 0 {
				l.aCount -= count
				delete(l.aHolders, txnID)
			}
			// Release all B-locks held by this transaction
			if count := l.bHolders[txnID]; count > 0 {
				l.bCount -= count
				delete(l.bHolders, txnID)
			}
			// Signal waiting goroutines that locks have been released
			l.cond.Broadcast()
			l.mu.Unlock()
		})
	}

	txn.OnDiscard(cleanup)
	txn.OnError(cleanup)
	txn.OnSuccess(cleanup)
}
