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

import "testing"

func Benchmark_InitialReadLock(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		txn := newTxn(1)
		lockSet := newLockSet[int]()

		b.StartTimer()
		lockSet.RLock(txn, 1)
	}
}

func Benchmark_SecondaryReadLock(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		txn := newTxn(1)
		lockSet := newLockSet[int]()

		// Perform the initial RLock, before starting the timer - we are not interested
		// in the cost of this in this test.
		lockSet.RLock(txn, 1)

		b.StartTimer()
		lockSet.RLock(txn, 1)
	}
}

func Benchmark_InitialWriteLock(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		txn := newTxn(1)
		lockSet := newLockSet[int]()

		b.StartTimer()
		lockSet.Lock(txn, 1)
	}
}

func Benchmark_SecondaryWriteLock(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		txn := newTxn(1)
		lockSet := newLockSet[int]()

		// Perform the initial Lock, before starting the timer - we are not interested
		// in the cost of this in this test.
		lockSet.Lock(txn, 1)

		b.StartTimer()
		lockSet.Lock(txn, 1)
	}
}

func Benchmark_ReadLockPromotion(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		txn := newTxn(1)
		lockSet := newLockSet[int]()

		// Perform the initial RLock, before starting the timer - we are not interested
		// in the cost of this in this test.
		lockSet.RLock(txn, 1)

		b.StartTimer()
		lockSet.Lock(txn, 1)
	}
}

// WARNING: This test tests internals, and so is highly disposable.
func Benchmark_ReadLockUnlock(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		txn := newTxn(1)
		lockSet := newLockSet[int]()

		lockSet.RLock(txn, 1)

		b.StartTimer()
		// In production, this will be called on transaction close - we could call
		// txn.Discard or similar, but we cannot isolate the unlocking cost if we do
		// so (we could bench it without the lock and subtract the cost though).
		lockSet.unlockAll(txn)
	}
}

// WARNING: This test tests internals, and so is highly disposable.
func Benchmark_WriteLockUnlock(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		txn := newTxn(1)
		lockSet := newLockSet[int]()

		lockSet.Lock(txn, 1)

		b.StartTimer()
		// In production, this will be called on transaction close - we could call
		// txn.Discard or similar, but we cannot isolate the unlocking cost if we do
		// so (we could bench it without the lock and subtract the cost though).
		lockSet.unlockAll(txn)
	}
}
