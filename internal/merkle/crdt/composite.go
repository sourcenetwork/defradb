// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package merklecrdt

import (
	"bytes"
	"context"
	"errors"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/internal/core"
	corecrdt "github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// MerkleCompositeDAG is a MerkleCRDT implementation of the CompositeDAG using MerkleClocks.
type MerkleCompositeDAG struct {
	store           datastore.DSReaderWriter
	key             keys.DataStoreKey
	schemaVersionID string
}

var _ core.ReplicatedData = (*MerkleCompositeDAG)(nil)

// NewMerkleCompositeDAG creates a new instance (or loaded from DB) of a MerkleCRDT
// backed by a CompositeDAG CRDT.
func NewMerkleCompositeDAG(
	store datastore.DSReaderWriter,
	schemaVersionID string,
	key keys.DataStoreKey,
) *MerkleCompositeDAG {
	return &MerkleCompositeDAG{
		store:           store,
		key:             key,
		schemaVersionID: schemaVersionID,
	}
}

func (m *MerkleCompositeDAG) HeadstorePrefix() keys.HeadstoreKey {
	return m.key.ToHeadStoreKey()
}

// DeleteDelta sets the values of CompositeDAG for a delete.
func (m *MerkleCompositeDAG) DeleteDelta() *corecrdt.CompositeDAGDelta {
	return &corecrdt.CompositeDAGDelta{
		DocID:           []byte(m.key.DocID),
		SchemaVersionID: m.schemaVersionID,
		Status:          client.Deleted,
	}
}

// Delta the value of the composite CRDT to DAG.
func (m *MerkleCompositeDAG) Delta() *corecrdt.CompositeDAGDelta {
	return &corecrdt.CompositeDAGDelta{
		DocID:           []byte(m.key.DocID),
		SchemaVersionID: m.schemaVersionID,
		Status:          client.Active,
	}
}

// Merge implements ReplicatedData interface.
// It ensures that the object marker exists for the given key.
// If it doesn't, it adds it to the store.
func (m *MerkleCompositeDAG) Merge(ctx context.Context, delta core.Delta) error {
	dagDelta, ok := delta.(*corecrdt.CompositeDAGDelta)
	if !ok {
		return corecrdt.ErrMismatchedMergeType
	}

	if dagDelta.Status.IsDeleted() {
		err := m.store.Set(ctx, m.key.ToPrimaryDataStoreKey().Bytes(), []byte{base.DeletedObjectMarker})
		if err != nil {
			return err
		}
		return m.deleteWithPrefix(ctx, m.key.WithValueFlag().WithFieldID(""))
	}

	// We cannot rely on the dagDelta.Status here as it may have been deleted locally, this is not
	// reflected in `dagDelta.Status` if sourced via P2P.  Updates synced via P2P should not undelete
	// the local representation of the document.
	versionKey := m.key.WithValueFlag().WithFieldID(keys.DATASTORE_DOC_VERSION_FIELD_ID)
	objectMarker, err := m.store.Get(ctx, m.key.ToPrimaryDataStoreKey().Bytes())
	hasObjectMarker := !errors.Is(err, corekv.ErrNotFound)
	if err != nil && hasObjectMarker {
		return err
	}

	if bytes.Equal(objectMarker, []byte{base.DeletedObjectMarker}) {
		versionKey = versionKey.WithDeletedFlag()
	}

	err = m.store.Set(ctx, versionKey.Bytes(), []byte(dagDelta.SchemaVersionID))
	if err != nil {
		return err
	}

	if !hasObjectMarker {
		// ensure object marker exists
		return m.store.Set(ctx, m.key.ToPrimaryDataStoreKey().Bytes(), []byte{base.ObjectMarker})
	}

	return nil
}

func (m MerkleCompositeDAG) deleteWithPrefix(ctx context.Context, key keys.DataStoreKey) error {
	iter, err := m.store.Iterator(ctx, corekv.IterOptions{
		Prefix: key.Bytes(),
	})
	if err != nil {
		return err
	}

	for {
		hasNext, err := iter.Next()
		if err != nil {
			return errors.Join(err, iter.Close())
		}
		if !hasNext {
			break
		}

		dsKey, err := keys.NewDataStoreKey(string(iter.Key()))
		if err != nil {
			return errors.Join(err, iter.Close())
		}

		if dsKey.InstanceType == keys.ValueKey {
			value, err := iter.Value()
			if err != nil {
				return errors.Join(err, iter.Close())
			}

			err = m.store.Set(ctx, dsKey.WithDeletedFlag().Bytes(), value)
			if err != nil {
				return errors.Join(err, iter.Close())
			}
		}

		err = m.store.Delete(ctx, dsKey.Bytes())
		if err != nil {
			return errors.Join(err, iter.Close())
		}
	}

	return iter.Close()
}
