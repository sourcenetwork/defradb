// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package vectorindex

import (
	"context"
	"encoding/binary"
	"math"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// PendingOp is the kind of change a pending entry records.
type PendingOp uint8

const (
	// PendingUpsert sets the node's vector, replacing any vector the graph holds for it.
	PendingUpsert PendingOp = iota + 1
	// PendingDelete removes the node from the results.
	PendingDelete
)

// PendingEntry is one write to a deferred index that has not been applied to the graph yet. There is
// at most one per node, so a later write to the same node overwrites (coalesces) an earlier one.
type PendingEntry struct {
	NodeID uint64
	Op     PendingOp
	// Vector is the new vector; only set for PendingUpsert.
	Vector []float32
}

// Encode returns the stored value: one op byte followed by the vector as little-endian float32s.
func (e PendingEntry) Encode() []byte {
	b := make([]byte, 1+4*len(e.Vector))
	b[0] = byte(e.Op)
	for i, f := range e.Vector {
		binary.LittleEndian.PutUint32(b[1+4*i:], math.Float32bits(f))
	}
	return b
}

// DecodePendingEntry parses a stored value for the given node.
func DecodePendingEntry(nodeID uint64, b []byte) (PendingEntry, error) {
	if len(b) < 1 || (len(b)-1)%4 != 0 {
		return PendingEntry{}, newErrMalformedPending(nodeID, len(b))
	}
	op := PendingOp(b[0])
	if op != PendingUpsert && op != PendingDelete {
		return PendingEntry{}, newErrUnknownPendingOp(nodeID, op)
	}
	e := PendingEntry{NodeID: nodeID, Op: op}
	if n := (len(b) - 1) / 4; n > 0 {
		e.Vector = make([]float32, n)
		for i := range e.Vector {
			e.Vector[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[1+4*i:]))
		}
	}
	return e, nil
}

// PutPending records entry as the node's pending write, replacing any earlier one.
func PutPending(ctx context.Context, collectionShortID, indexID, epoch uint32, entry PendingEntry) error {
	txn := datastore.CtxMustGetTxn(ctx)
	key := keys.NewVectorPendingKey(collectionShortID, indexID, epoch, entry.NodeID)
	if err := txn.Datastore().Set(ctx, &key, entry.Encode()); err != nil {
		return newErrVectorIndexStore(err, collectionShortID, indexID, epoch)
	}
	return nil
}

// DeletePending removes the node's pending entry, if any.
func DeletePending(ctx context.Context, collectionShortID, indexID, epoch uint32, nodeID uint64) error {
	txn := datastore.CtxMustGetTxn(ctx)
	key := keys.NewVectorPendingKey(collectionShortID, indexID, epoch, nodeID)
	if err := txn.Datastore().Delete(ctx, &key); err != nil && !errors.Is(err, corekv.ErrNotFound) {
		return newErrVectorIndexStore(err, collectionShortID, indexID, epoch)
	}
	return nil
}

// ListPending returns up to limit pending entries of the epoch in node id order, reading through the
// transaction on ctx. A limit of zero or less returns all of them.
func ListPending(
	ctx context.Context,
	collectionShortID, indexID, epoch uint32,
	limit int,
) ([]PendingEntry, error) {
	txn := datastore.CtxMustGetTxn(ctx)
	prefix := keys.NewVectorPendingPrefix(collectionShortID, indexID, epoch)

	iter, err := txn.Datastore().Iterator(ctx, datastore.IterOptions{Prefix: &prefix})
	if err != nil {
		return nil, newErrVectorIndexStore(err, collectionShortID, indexID, epoch)
	}

	var entries []PendingEntry
	for limit <= 0 || len(entries) < limit {
		hasNext, err := iter.Next()
		if err != nil {
			return nil, newErrVectorIndexStore(errors.Join(err, iter.Close()), collectionShortID, indexID, epoch)
		}
		if !hasNext {
			break
		}

		nodeID, err := keys.DecodeVectorPendingNodeID(iter.Key())
		if err != nil {
			return nil, newErrVectorIndexStore(errors.Join(err, iter.Close()), collectionShortID, indexID, epoch)
		}
		val, err := iter.Value()
		if err != nil {
			return nil, newErrVectorIndexStore(errors.Join(err, iter.Close()), collectionShortID, indexID, epoch)
		}
		entry, err := DecodePendingEntry(nodeID, val)
		if err != nil {
			return nil, newErrVectorIndexStore(errors.Join(err, iter.Close()), collectionShortID, indexID, epoch)
		}
		entries = append(entries, entry)
	}

	if err := iter.Close(); err != nil {
		return nil, newErrVectorIndexStore(err, collectionShortID, indexID, epoch)
	}
	return entries, nil
}
