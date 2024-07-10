// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package errors

import "errors"

var (
	// ErrConflict is returned when a transaction conflicts with another transaction. This can
	// happen if the read rows had been updated concurrently by another transaction.
	//
	// This error is copied from badger because it is used in a few places outside of the datastore
	// package. This copy ensures that we can build in environments that do not support badger.
	//
	//nolint:revive
	ErrTxnConflict = errors.New("Transaction Conflict. Please retry")
	// ErrDiscardedTxn is returned if a previously discarded transaction is re-used.
	//
	// This error is copied from badger because it is used in a few places outside of the datastore
	// package. This copy ensures that we can build in environments that do not support badger.
	//
	//nolint:revive
	ErrTxnDiscarded = errors.New("This transaction has been discarded. Create a new one")
	// ErrReadOnlyTxn is returned if an update function is called on a read-only transaction.
	//
	// This error is copied from badger because it is used in a few places outside of the datastore
	// package. This copy ensures that we can build in environments that do not support badger.
	//
	//nolint:revive
	ErrReadOnlyTxn = errors.New("No sets or deletes are allowed in a read-only transaction")
)
