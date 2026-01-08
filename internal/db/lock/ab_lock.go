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
// LockA will prevent LockB from being acquired, but it will not prevent
// other LockAs from being acquired, and vice versa.
//
// Acquired locks are released on transaction close.
type abLock struct {
	aGroup sync.WaitGroup
	bGroup sync.WaitGroup
}

func (l *abLock) LockA(txn txn) {
	l.aGroup.Add(1)

	// We need to make sure this is only ever called once!  It is permitted to discard
	// a transaction after it is commited whether it errors or not.
	var once sync.Once
	done := func() { once.Do(func() { l.aGroup.Done() }) }
	txn.OnDiscard(done)
	txn.OnError(done)
	txn.OnSuccess(done)

	l.bGroup.Wait()
}

func (l *abLock) LockB(txn txn) {
	l.bGroup.Add(1)

	// We need to make sure this is only ever called once!  It is permitted to discard
	// a transaction after it is commited whether it errors or not.
	var once sync.Once
	done := func() { once.Do(func() { l.bGroup.Done() }) }
	txn.OnDiscard(done)
	txn.OnError(done)
	txn.OnSuccess(done)

	l.aGroup.Wait()
}
