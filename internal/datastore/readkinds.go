// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package datastore

import (
	"strings"
	"sync/atomic"
)

// ReadKind names one class of key a transaction's read set can hold.
//
// The underlying store reports a commit conflict without naming the contended key, so the
// set of classes a transaction read is the only lead available for working out what it
// contended with. The classes are bit flags so a whole transaction's set fits in one word
// and recording a read costs an atomic OR with no allocation.
type ReadKind uint32

const (
	// ReadDoc is a document field value or its priority marker, under the datastore.
	ReadDoc ReadKind = 1 << iota
	// ReadIndex is a secondary index entry, under the datastore. Set by the index write
	// path, which is the only place that knows an encoded datastore key is an index entry.
	ReadIndex
	// ReadUniqueIndex is the existence check a unique index performs before writing its
	// entry. Set by the unique index write path for the same reason as ReadIndex.
	ReadUniqueIndex
	// ReadHead is a headstore key: a document's or collection's current DAG heads.
	ReadHead
	// ReadBlock is a blockstore key: an IPLD block or its to-merge marker.
	ReadBlock
	// ReadSystem is a systemstore key that is not a sequence.
	ReadSystem
	// ReadSequence is a systemstore /seq/ key. Short ID allocation commits the sequence in
	// its own transaction, so a merge never carries one in its read set and this stays zero
	// on that path. It is kept for callers that do read a sequence inline.
	ReadSequence
	// ReadPeer is a peerstore key.
	ReadPeer
	// ReadEnc is an encryption keystore key.
	ReadEnc
	// ReadRangeIterator means the transaction opened an iterator bounded by start and end
	// rather than by prefix. Badger's prefix option is unset for those, so its bounds
	// check reads the first key past the range into the read set.
	ReadRangeIterator
	// ReadPrefixIterator means the transaction opened a prefix-bounded iterator, which
	// cannot read past its range.
	ReadPrefixIterator
)

// readKindNames is ordered to match the bit order above.
var readKindNames = [...]string{
	"doc",
	"index",
	"uniqueIndex",
	"head",
	"block",
	"system",
	"sequence",
	"peer",
	"enc",
	"rangeIter",
	"prefixIter",
}

// ReadKindCount is the number of distinct kinds, for callers keeping a counter per kind.
const ReadKindCount = len(readKindNames)

// ReadKindName returns the name of the kind held in bit i, or "" if i is out of range.
func ReadKindName(i int) string {
	if i < 0 || i >= len(readKindNames) {
		return ""
	}
	return readKindNames[i]
}

// String renders the set as a pipe-separated list in bit order, e.g. "doc|head|rangeIter".
// An empty set renders as "none".
func (k ReadKind) String() string {
	if k == 0 {
		return "none"
	}
	var b strings.Builder
	for i, name := range readKindNames {
		if k&(1<<uint(i)) == 0 {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('|')
		}
		b.WriteString(name)
	}
	return b.String()
}

// ReadKinds accumulates the kinds of key a single transaction has read.
//
// The zero value is ready to use. It is safe for concurrent use, which matters because a
// transaction may be read from more than one goroutine even though a merge is not.
type ReadKinds struct {
	bits atomic.Uint32
}

// Mark records that a key of the given kind entered the read set. Safe on a nil receiver
// so call sites do not have to check whether the transaction records kinds.
func (r *ReadKinds) Mark(kind ReadKind) {
	if r == nil {
		return
	}
	r.bits.Or(uint32(kind))
}

// Kinds returns the set recorded so far.
func (r *ReadKinds) Kinds() ReadKind {
	if r == nil {
		return 0
	}
	return ReadKind(r.bits.Load())
}

// readKindsCarrier is implemented by transactions that record read kinds. It is optional:
// shims and mocks that do not record them are simply not asked.
type readKindsCarrier interface {
	ReadKinds() *ReadKinds
}

// ReadKindsOf returns the recorder for txn, or nil if txn does not record read kinds.
// A nil result is usable: every method on ReadKinds tolerates it.
//
// Used by the few call sites that know a key's class from its Go type rather than its
// encoded bytes: an index entry and a document value are both datastore keys and are
// indistinguishable once encoded.
func ReadKindsOf(txn any) *ReadKinds {
	if c, ok := txn.(readKindsCarrier); ok {
		return c.ReadKinds()
	}
	return nil
}
