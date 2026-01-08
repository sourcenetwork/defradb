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

// LockSet manages a set of available locks.
type LockSet struct {
	collectionLockSet *lockSet[uint32]
}

// NewLockSet creates a new LockSet that manages a set of mutexes.
//
// The returned instance is completely independant from any other
// existing LockSet instances.
func NewLockSet() *LockSet {
	return &LockSet{
		collectionLockSet: newLockSet[uint32](),
	}
}

// CollectionLock acquires a write lock for the given collection short id.
//
// This will prevent all other transactions from acquiring a read or write lock
// to the given collection until the lock is released.  The lock will be released
// when the transaction is either committed or discarded.
//
// The acquired lock will not block other threads operating within this transaction.
func (l *LockSet) CollectionLock(txn txn, collectionShortID uint32) {
	l.collectionLockSet.Lock(txn, collectionShortID)
}

// CollectionRLock acquires a read lock for the given collection short id.
//
// This will prevent all other transactions from acquiring a write lock
// to the given collection until the lock is released.  The lock will be released
// when the transaction is either committed or discarded.
//
// The read lock can be promoted to a write lock by this transaction, however, currently,
// it does this by first releasing the read lock and then acquiring a write lock.  This can
// permit competing transaction-locks to acquire a write lock, blocking this thread's acquisition
// of the write lock, and allowing both the other transaction's thread, and any previously
// read-locked threads for this transaction to progress concurrently.
func (l *LockSet) CollectionRLock(txn txn, collectionShortID uint32) {
	l.collectionLockSet.RLock(txn, collectionShortID)
}

func (l *LockSet) RLockAll(txn txn) {
	l.collectionLockSet.RLockAll(txn)
}
