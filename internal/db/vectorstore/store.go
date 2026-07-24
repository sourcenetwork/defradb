// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package vectorstore binds the datastore-independent HNSW engine (internal/index/hnsw) to
// DefraDB's transaction datastore. It is the one place that knows how a vector index's graph is
// laid out in the key/value store, so both the write path (package db, maintaining the graph as
// documents change) and the read path (the query planner, searching the graph) build a graph the
// same way without either importing the other.
package vectorstore

import (
	"context"
	"errors"

	"github.com/sourcenetwork/corekv"

	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/index/hnsw"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// NodeStore implements hnsw.NodeStore over one vector index's (collection, index, epoch) keyspace.
// It holds ctx because, like the rest of the index code, it resolves the transaction from ctx per
// call rather than pinning one.
type NodeStore struct {
	ctx               context.Context
	collectionShortID uint32
	indexID           uint32
	epoch             uint32
}

var _ hnsw.NodeStore = (*NodeStore)(nil)

// NewNodeStore returns a NodeStore scoped to the given (collection, index, epoch) keyspace.
func NewNodeStore(ctx context.Context, collectionShortID, indexID, epoch uint32) *NodeStore {
	return &NodeStore{
		ctx:               ctx,
		collectionShortID: collectionShortID,
		indexID:           indexID,
		epoch:             epoch,
	}
}

// GetNode returns the node with the given id, or ok == false if no such node has been stored.
func (s *NodeStore) GetNode(id hnsw.NodeID) (hnsw.Node, bool, error) {
	txn := datastore.CtxMustGetTxn(s.ctx)
	key := keys.NewVectorNodeKey(s.collectionShortID, s.indexID, s.epoch, uint64(id))

	val, err := txn.Datastore().Get(s.ctx, &key)
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return hnsw.Node{}, false, nil
		}
		return hnsw.Node{}, false, newErrVectorIndexStore(err)
	}

	node, err := hnsw.UnmarshalNode(val)
	if err != nil {
		return hnsw.Node{}, false, newErrVectorIndexStore(err)
	}
	return node, true, nil
}

// PutNode stores n, replacing any previously stored node with the same id.
func (s *NodeStore) PutNode(n hnsw.Node) error {
	txn := datastore.CtxMustGetTxn(s.ctx)
	key := keys.NewVectorNodeKey(s.collectionShortID, s.indexID, s.epoch, uint64(n.ID))

	val, err := hnsw.MarshalNode(n)
	if err != nil {
		return newErrVectorIndexStore(err)
	}

	if err := txn.Datastore().Set(s.ctx, &key, val); err != nil {
		return newErrVectorIndexStore(err)
	}
	return nil
}

// GetMeta returns the graph's meta singleton, or a zero-value Meta with Empty set to true if no
// graph has been built yet.
func (s *NodeStore) GetMeta() (hnsw.Meta, error) {
	txn := datastore.CtxMustGetTxn(s.ctx)
	key := keys.NewVectorMetaKey(s.collectionShortID, s.indexID, s.epoch)

	val, err := txn.Datastore().Get(s.ctx, &key)
	if err != nil {
		if errors.Is(err, corekv.ErrNotFound) {
			return hnsw.Meta{Empty: true}, nil
		}
		return hnsw.Meta{}, newErrVectorIndexStore(err)
	}

	meta, err := hnsw.UnmarshalMeta(val)
	if err != nil {
		return hnsw.Meta{}, newErrVectorIndexStore(err)
	}
	return meta, nil
}

// PutMeta stores m as the graph's meta singleton, replacing any previously stored value.
func (s *NodeStore) PutMeta(m hnsw.Meta) error {
	txn := datastore.CtxMustGetTxn(s.ctx)
	key := keys.NewVectorMetaKey(s.collectionShortID, s.indexID, s.epoch)

	val, err := hnsw.MarshalMeta(m)
	if err != nil {
		return newErrVectorIndexStore(err)
	}

	if err := txn.Datastore().Set(s.ctx, &key, val); err != nil {
		return newErrVectorIndexStore(err)
	}
	return nil
}

// IterateNodes iterates over all non-deleted node entries stored for this (collection, index,
// epoch), calling fn for each. Iteration stops early if fn returns an error, and that error is
// returned to the caller.
func (s *NodeStore) IterateNodes(fn func(hnsw.Node) error) error {
	txn := datastore.CtxMustGetTxn(s.ctx)
	prefix := keys.NewVectorNodePrefix(s.collectionShortID, s.indexID, s.epoch)

	iter, err := txn.Datastore().Iterator(s.ctx, datastore.IterOptions{Prefix: &prefix})
	if err != nil {
		return newErrVectorIndexStore(err)
	}

	for {
		hasNext, err := iter.Next()
		if err != nil {
			return newErrVectorIndexStore(errors.Join(err, iter.Close()))
		}
		if !hasNext {
			break
		}

		val, err := iter.Value()
		if err != nil {
			return newErrVectorIndexStore(errors.Join(err, iter.Close()))
		}

		node, err := hnsw.UnmarshalNode(val)
		if err != nil {
			return newErrVectorIndexStore(errors.Join(err, iter.Close()))
		}
		if node.Deleted {
			continue
		}

		if err := fn(node); err != nil {
			return errors.Join(err, iter.Close())
		}
	}

	if err := iter.Close(); err != nil {
		return newErrVectorIndexStore(err)
	}
	return nil
}
