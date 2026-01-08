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

import "sync"

// txnLocks manages a set of mutexes scoped to each transaction. Each transaction will have zero or one
// mutexes associated with it.
//
// Acquired locks are not automatically unlocked, it is the responsibility of consumers of this type
// to unlock any locks that they have acquired.
//
// This type deletes the mutex associated with each transaction when the transaction is commited/discarded.
//
// Only public functions on this type should be called from outside of this type.  Calling private members
// from anywhere else, internal or otherwise, removes any guarantees over its state's correctness.
type txnLocks struct {
	// locksByTxnID is a map of RWMutexes by transaction ID.
	//
	// Hosting this inside a dedicated object instead of on the transaction object itself, allows the
	// scope to be significantly shrunk.
	//
	// For example, an instance of an `txnLocks` object could be a property on a `lockSet` scoped to datastore
	// collection keys, and used to ensure that the management of locks within that `lockset` is essentially
	// single threaded per transaction, whilst allowing the bulk of Defra operation-processing to remain concurrent
	// for the transaction.
	locksByTxnID     map[uint64]*sync.RWMutex
	locksByTxnIDLock sync.Mutex
}

func newTxnLocks() *txnLocks {
	return &txnLocks{
		locksByTxnID: map[uint64]*sync.RWMutex{},
	}
}

func (l *txnLocks) RLock(txn txn) {
	lock := l.getLock(txn)
	lock.RLock()
}

func (l *txnLocks) RUnlock(txn txn) {
	lock := l.getLock(txn)
	lock.RUnlock()
}

func (l *txnLocks) Lock(txn txn) {
	lock := l.getLock(txn)
	lock.Lock()
}

func (l *txnLocks) Unlock(txn txn) {
	lock := l.getLock(txn)
	lock.Unlock()
}

func (l *txnLocks) getLock(txn txn) *sync.RWMutex {
	id := txn.ID()

	l.locksByTxnIDLock.Lock()
	lock, ok := l.locksByTxnID[id]
	if !ok {
		lock = &sync.RWMutex{}
		l.locksByTxnID[id] = lock

		txn.OnDiscard(func() { l.deleteMutex(id) })
		txn.OnError(func() { l.deleteMutex(id) })
		txn.OnSuccess(func() { l.deleteMutex(id) })
	}
	// WARNING - it is very important to unlock this before using the returned lock,
	// otherwise if the returned lock is locked, this locksLock will remain locked
	// until the returned lock is finally acquired - defeating the purpose of any smaller
	// scope locks, or read locks, managed behind this one.
	l.locksByTxnIDLock.Unlock()

	return lock
}

func (l *txnLocks) deleteMutex(id uint64) {
	l.locksByTxnIDLock.Lock()
	defer l.locksByTxnIDLock.Unlock()

	// Unlocking the mutex is managed explicitly by other code, we should not unlock it here.
	// It is safe to assume that it is unlocked before this line is called - if it is not, then
	// there is a bug elsewhere.
	delete(l.locksByTxnID, id)
}
