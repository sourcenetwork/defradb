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

// lockSet manages Defra-wide locks, scoped to an arbitrary Defra-element delimited by the given `TKey`
// parameters, such as a collection, or document id.
//
// Calls into this type made by the same transaction for the same key will not compete for the same lock.
//
// A single lock is only ever held by any given transaction.  Subsequent calls to RLock or Lock will be
// ignored if the corresponding lock is already held by the transaction. If a read lock is held by the
// transaction when a write lock is requested, the read lock will be promoted to a write lock.
//
// All locks held by the transaction will be released when the transaction closes (on commit/discard).
// The current code relies on this, so be very careful if allowing explicit unlocking of keys.  It is also
// in keeping with other systems - it seems very unlikely that locks *should* ever be managed independently
// from their transactions, as if a transaction desires a different lock scope, then it should probably be
// doing whatever it is doing in multiple transactions, or without a transaction, instead.
//
// If a requested lock is already held, the requesting transaction will be blocked from acquiring any new
// locks, and from unlocking held locks, until the requested lock becomes available.  For more info on this,
// see the documentation for the `txnLocks` member of this type.
//
// Only public functions on this type should be called from outside of this type.  Calling private members
// from anywhere else, internal or otherwise, removes any guarantees over its state's correctness.
type lockSet[TKey comparable] struct {
	// locksByKey contains the core mutexes managed by `lockSet`, indexed by the consumer provided keys.
	//
	// All other state held by `lockset` is to support access to this property, all other mutexes exist
	// either to manage internal access to this property, or are pointers to the mutexes within this set.
	locksByKey map[TKey]*sync.RWMutex
	// locksByKeyLock gates access to `locksByKey`.
	//
	// The number of write-lock calls to this could be considerably reduced if `locksByKey` is
	// only mutated when new keys are added locally (e.g. collections via `AddCollection`).  It would however
	// introduce a new, potentialy very damaging, way in which bugs could be introduced (if `AddCollection`
	//  and friends dont add to this).
	locksByKeyLock sync.RWMutex

	// heldRLocksByTxnID is a map of a map of pointers to the mutexes contained within `locksByKey`.
	//
	// It only contains key-values to read locks acquired by the transaction-key pair.
	//
	// By mapping to the key-lock pointer here, we avoid the need to hold `locksByKeyLock`
	// when accessing it.
	heldRLocksByTxnID map[uint64]map[TKey]*sync.RWMutex
	txnRLockLock      sync.RWMutex

	// heldLocksByTxnID is a map of a map of pointers to the mutexes contained within `locksByKey`.
	//
	// It only contains key-values to write locks acquired by the transaction-key pair.
	//
	// By mapping to the key-lock pointer here, we avoid the need to hold `locksByKeyLock`
	// when accessing it.
	heldLocksByTxnID map[uint64]map[TKey]*sync.RWMutex
	txnLockLock      sync.RWMutex

	// allLock allows all writes to progress so long as there are no concurrent reads,
	// and all reads to progress so long as there are no writes.
	//
	// Read locking is handled by 'B' locks - this is done when acquiring an RLockAll, which
	// must prevent writes of any kind, including for new keys that may not have existed when the
	// B lock was first acquired.
	//
	// Write locking is handled by 'A' locks - this is done when acquiring a write lock for any key.
	// Locking within the keys are handled by other means within this type, but a write lock of any
	// key must block an RLockAll from being acquired.
	allLock abLock

	// txnLocks manages transaction level locks scoped to this lockSet.
	//
	// It is important that operations within the lockset are not concurrently executed for the same transaction -
	// every public function on this type has a bunch of race conditions that can occur if they were to concurrently
	// process for the same transaction, and it is worth the performance loss to just lock them all down per transaction
	// right now.  Doing this also allows us to reduce the scope of some cross-transaction locks, as we are now free
	// to assume that each transaction will not be processing multiple locks at the same time - more than offsetting
	// the non-concurrent intra-transaction performance loss in most use cases.
	//
	// This does not prevent concurrent Defra operations for a given transaction, it only locks the locking
	// and unlocking of locks within this lockset's scope (e.g. collections) - if the lockset locks are free to acquire -
	// if acquiring the sought-after lock (e.g. the collection lock) is blocked, e.g. by a competing write-lock for the
	// collection, then other concurrent operations on the given transaction will be blocked from acquiring their
	// `txnLocks`lock - preventing those threads from progressing until the lockset lock (e.g. collection lock) is
	// acquired.
	txnLocks *txnLocks
}

func newLockSet[TKey comparable]() *lockSet[TKey] {
	return &lockSet[TKey]{
		locksByKey:        map[TKey]*sync.RWMutex{},
		heldRLocksByTxnID: map[uint64]map[TKey]*sync.RWMutex{},
		heldLocksByTxnID:  map[uint64]map[TKey]*sync.RWMutex{},
		txnLocks:          newTxnLocks(),
	}
}

// Lock blocks until the given transaction has managed to acquire the write lock for the given key.
//
// If the transaction already has the write lock, this will no-op.
//
// If the transaction already has a read lock, it will promote the read lock into a write lock.
// WARNING: It currently does this by unlocking the existing read lock, and then acquiring the write lock -
// this can unblock other threads trying to acquire the lock, as well as the thread holding the original
// read lock.
//
// If a lock was acquired, it will be unlocked on transaction commit/discard.
func (l *lockSet[TKey]) Lock(txn txn, key TKey) {
	l.txnLocks.Lock(txn)
	defer l.txnLocks.Unlock(txn)

	if l.hasLock(txn, key) {
		return
	}

	var lock *sync.RWMutex
	if l.hasRLock(txn, key) {
		// From an internal perspective this code appears dead, however if a user submits two
		// concurrent operations for the same transaction, with the first acquiring a read lock,
		// and the second requesting a write lock, this code path may be reached.
		l.locksByKeyLock.RLock()
		lock = l.locksByKey[key]
		l.locksByKeyLock.RUnlock()

		l.txnRLockLock.Lock()
		delete(l.heldRLocksByTxnID[txn.ID()], key)

		// If *this* transaction already holds a read lock, we must unlock it in order to avoid a deadlock.
		//
		// WARNING - This introduces a race condition between the RUnlock and Lock calls. This is easily
		// avoided in internal code by ensuring that the write lock is acquired before any read locks.
		//
		// However, if a user submits two concurrent operations for the same transaction, with the first
		// acquiring a read lock, and the second requesting a write lock, this code path, and the race
		// condition may be reached.
		//
		// We could consider doing something similar to https://upstash.com/blog/upgradable-rwlock-for-go in
		// this case if we really want (the correctness of the code in this blog has not been verified by us).
		lock.RUnlock()
		l.txnRLockLock.Unlock()
	} else {
		// A write lock must be held as we need to write to the set if the
		// lock key does not yet exist.  The lock cannot be unlocked
		// until after the potential write without introducing a race condition.
		l.locksByKeyLock.Lock()
		var ok bool
		lock, ok = l.locksByKey[key]

		if !ok {
			lock = &sync.RWMutex{}
			// If the lock key does not exist yet, we must add it so
			// that the read lock can be held if a write lock is attempted.
			l.locksByKey[key] = lock
		}
		l.locksByKeyLock.Unlock()
	}

	lock.Lock()
	// Block RLockAll from acquiring a lock until this write lock has completed, without blocking
	// write locks to other keys.
	l.allLock.LockA(txn)

	l.txnLockLock.Lock()

	txnID := txn.ID()
	txnLocks, ok := l.heldLocksByTxnID[txnID]
	if !ok {
		txnLocks = map[TKey]*sync.RWMutex{}
		l.heldLocksByTxnID[txnID] = txnLocks
	}
	// `l.txnLocks` protects against concurrent mutations to `txnLocks`, so we can unlock
	// this lock as soon as we have finished any potential mutations to the outer map
	// in `l.heldLocksByTxnID`.
	l.txnLockLock.Unlock()

	if !ok {
		// If this is the first write lock held by this transaction,
		// add an unlock-all call to execute when the txn closes.
		//
		// It doesn't really matter if `UnlockAll` is called multiple times, it is
		// just a little wasteful.
		txn.OnDiscard(func() { l.unlockAll(txn) })
		txn.OnError(func() { l.unlockAll(txn) })
		txn.OnSuccess(func() { l.unlockAll(txn) })
	}

	txnLocks[key] = lock
}

// Lock blocks until the given transaction has managed to acquire the read lock for the given key.
//
// If the transaction already has either a read or write lock, this will no-op.
//
// If a lock was acquired, it will be unlocked on transaction commit/discard.
func (l *lockSet[TKey]) RLock(txn txn, key TKey) {
	l.txnLocks.Lock(txn)
	defer l.txnLocks.Unlock(txn)

	if l.hasAnyLock(txn, key) {
		// Write lock documentation:
		//
		// We need to permit RLock-gated operations from the thread holding the write lock,
		// so if this context holds the write lock we must not RLock (it will deadlock in most
		// situations).
		//
		// We need this RLock-skipping behaviour to allow the wlocking thread to perform rlocking
		// actions without deadlocking, e.g. a `truncate` call might also read data.
		//
		// This behaviour is thread-safe and it is desirable to bypass the RLock in any child
		// threads of the write-lock holding thread.  The transaction-conflict protections will
		// still apply in such a case.
		//
		// -------
		//
		// Read lock documentation:
		//
		// It is safe to skip the RLock, as any held RLocks will only ever be released upon close
		// of the transaction anyway - there is no risk of the original RLock being unlocked before
		// this thread/action has been completed.
		//
		// We know that the RLock was successfully acquired, as otherwise we would not have been able
		// to detect it via the `l.hasAnyLock` call - so there is no risk of this thread progressing
		// whilst the original RLocked thread is blocked waiting to acquire the lock.
		//
		// Thanks to the `l.txnLocks` call at the top of this RLock function, we can rule out possible
		// race conditions in multiple concurrent actions for the same transaction acquiring read locks
		// at roughly the same time.
		return
	}

	// A write lock must be held as we need to write to the set if the
	// lock key does not yet exist.  The lock cannot be unlocked
	// until after the potential write without introducing a race condition.
	l.locksByKeyLock.Lock()
	lock, ok := l.locksByKey[key]

	if !ok {
		lock = &sync.RWMutex{}
		// If the lock key does not exist yet, we must add it so
		// that the read lock can be held if a write lock is attempted.
		l.locksByKey[key] = lock
	}
	l.locksByKeyLock.Unlock()

	lock.RLock()

	l.txnRLockLock.Lock()

	txnID := txn.ID()
	txnRLocks, ok := l.heldRLocksByTxnID[txnID]
	if !ok {
		txnRLocks = map[TKey]*sync.RWMutex{}
		l.heldRLocksByTxnID[txnID] = txnRLocks
	}
	// `l.txnLocks` protects against concurrent mutations to `txnRLocks`, so we can unlock
	// this lock as soon as we have finished any potential mutations to the outer map
	// in `l.heldRLocksByTxnID`.
	l.txnRLockLock.Unlock()

	if !ok {
		// If this is the first rlock held by this transaction,
		// add an unlock-all call to execute when the txn closes.
		//
		// It doesn't really matter if `RUnlockAll` is called multiple times, it is
		// just a little wasteful.
		txn.OnDiscard(func() { l.rUnlockAll(txn) })
		txn.OnError(func() { l.rUnlockAll(txn) })
		txn.OnSuccess(func() { l.rUnlockAll(txn) })
	}

	txnRLocks[key] = lock
}

func (l *lockSet[TKey]) RLockAll(txn txn) {
	l.allLock.LockB(txn)
}

func (l *lockSet[TKey]) rUnlockAll(txn txn) {
	l.txnLocks.RLock(txn)
	defer l.txnLocks.RUnlock(txn)

	txnID := txn.ID()

	l.txnRLockLock.Lock()
	locks, ok := l.heldRLocksByTxnID[txnID]
	delete(l.heldRLocksByTxnID, txnID)
	l.txnRLockLock.Unlock()

	if ok {
		for _, lock := range locks {
			lock.RUnlock()
		}
	}
}

func (l *lockSet[TKey]) unlockAll(txn txn) {
	l.txnLocks.RLock(txn)
	defer l.txnLocks.RUnlock(txn)

	txnID := txn.ID()

	l.txnLockLock.Lock()
	locks, ok := l.heldLocksByTxnID[txnID]
	delete(l.heldLocksByTxnID, txnID)
	l.txnLockLock.Unlock()

	if ok {
		for _, lock := range locks {
			lock.Unlock()
		}
	}
}

func (l *lockSet[TKey]) hasAnyLock(txn txn, key TKey) bool {
	if l.hasRLock(txn, key) {
		return true
	}

	return l.hasLock(txn, key)
}

func (l *lockSet[TKey]) hasRLock(txn txn, key TKey) bool {
	txnID := txn.ID()

	l.txnRLockLock.RLock()
	txnRLocks, ok := l.heldRLocksByTxnID[txnID]
	l.txnRLockLock.RUnlock()

	if ok {
		// Because all callers of this `hasRLock` function have locked per transaction, we can safely read from
		// txnRLocks without an additional lock.
		if _, ok := txnRLocks[key]; ok {
			return ok
		}
	}

	return false
}

func (l *lockSet[TKey]) hasLock(txn txn, key TKey) bool {
	txnID := txn.ID()

	l.txnLockLock.RLock()
	txnLocks, ok := l.heldLocksByTxnID[txnID]
	l.txnLockLock.RUnlock()

	if ok {
		// Because all callers of this `hasLock` function have locked per transaction, we can safely read from
		// txnLocks without an additional lock.
		if _, ok := txnLocks[key]; ok {
			return ok
		}
	}

	return false
}
