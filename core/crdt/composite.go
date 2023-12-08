// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crdt

import (
	"bytes"
	"context"
	"sort"
	"strings"

	dag "github.com/ipfs/boxo/ipld/merkledag"
	ds "github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/query"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ugorji/go/codec"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/errors"
)

// CompositeDAGDelta represents a delta-state update made of sub-MerkleCRDTs.
type CompositeDAGDelta struct {
	DocID     []byte
	FieldName string
	Priority  uint64
	// SchemaVersionID is the schema version datastore key at the time of commit.
	//
	// It can be used to identify the collection datastructure state at the time of commit.
	SchemaVersionID string
	// Status represents the status of the document. By default it is `Active`.
	// Alternatively, if can be set to `Deleted`.
	Status  client.DocumentStatus
	SubDAGs []core.DAGLink `json:"-"` // should not be marshalled
}

var _ core.CompositeDelta = (*CompositeDAGDelta)(nil)

// GetPriority gets the current priority for this delta.
func (delta *CompositeDAGDelta) GetPriority() uint64 {
	return delta.Priority
}

// SetPriority will set the priority for this delta.
func (delta *CompositeDAGDelta) SetPriority(prio uint64) {
	delta.Priority = prio
}

// Marshal will serialize this delta to a byte array.
func (delta *CompositeDAGDelta) Marshal() ([]byte, error) {
	h := &codec.CborHandle{}
	buf := bytes.NewBuffer(nil)
	enc := codec.NewEncoder(buf, h)
	err := enc.Encode(delta)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Links returns the links for this delta.
func (delta *CompositeDAGDelta) Links() []core.DAGLink {
	return delta.SubDAGs
}

// CompositeDAG is a CRDT structure that is used to track a collection of sub MerkleCRDTs.
type CompositeDAG struct {
	baseCRDT
}

var _ core.ReplicatedData = (*CompositeDAG)(nil)

func NewCompositeDAG(
	store datastore.DSReaderWriter,
	schemaVersionKey core.CollectionSchemaVersionKey,
	key core.DataStoreKey,
	fieldName string,
) CompositeDAG {
	return CompositeDAG{newBaseCRDT(store, key, schemaVersionKey, fieldName)}
}

// Value is a no-op for a CompositeDAG.
func (c CompositeDAG) Value(ctx context.Context) ([]byte, error) {
	return nil, nil
}

// Set applies a delta to the composite DAG CRDT. TBD
func (c CompositeDAG) Set(links []core.DAGLink) *CompositeDAGDelta {
	// make sure the links are sorted lexicographically by CID
	sort.Slice(links, func(i, j int) bool {
		return strings.Compare(links[i].Cid.String(), links[j].Cid.String()) < 0
	})
	return &CompositeDAGDelta{
		DocID:           []byte(c.key.DocID),
		FieldName:       c.fieldName,
		SchemaVersionID: c.schemaVersionKey.SchemaVersionId,
		SubDAGs:         links,
	}
}

// Merge implements ReplicatedData interface.
// It ensures that the object marker exists for the given key.
// If it doesn't, it adds it to the store.
func (c CompositeDAG) Merge(ctx context.Context, delta core.Delta) error {
	dagDelta, isDagDelta := delta.(*CompositeDAGDelta)

	if isDagDelta && dagDelta.Status.IsDeleted() {
		err := c.store.Put(ctx, c.key.ToPrimaryDataStoreKey().ToDS(), []byte{base.DeletedObjectMarker})
		if err != nil {
			return err
		}
		return c.deleteWithPrefix(ctx, c.key.WithValueFlag().WithFieldId(""))
	}

	// We cannot rely on the dagDelta.Status here as it may have been deleted locally, this is not
	// reflected in `dagDelta.Status` if sourced via P2P.  Updates synced via P2P should not undelete
	// the local reperesentation of the document.
	versionKey := c.key.WithValueFlag().WithFieldId(core.DATASTORE_DOC_VERSION_FIELD_ID)
	objectMarker, err := c.store.Get(ctx, c.key.ToPrimaryDataStoreKey().ToDS())
	hasObjectMarker := !errors.Is(err, ds.ErrNotFound)
	if err != nil && hasObjectMarker {
		return err
	}

	if bytes.Equal(objectMarker, []byte{base.DeletedObjectMarker}) {
		versionKey = versionKey.WithDeletedFlag()
	}

	var schemaVersionId string
	if isDagDelta {
		// If this is a CompositeDAGDelta take the datastore schema version from there.
		// This is particularly important for P2P synced dags, as they may arrive here without having
		// been migrated yet locally.
		schemaVersionId = dagDelta.SchemaVersionID
	} else {
		schemaVersionId = c.schemaVersionKey.SchemaVersionId
	}

	err = c.store.Put(ctx, versionKey.ToDS(), []byte(schemaVersionId))
	if err != nil {
		return err
	}

	if !hasObjectMarker {
		// ensure object marker exists
		return c.store.Put(ctx, c.key.ToPrimaryDataStoreKey().ToDS(), []byte{base.ObjectMarker})
	}

	return nil
}

func (c CompositeDAG) deleteWithPrefix(ctx context.Context, key core.DataStoreKey) error {
	q := query.Query{
		Prefix: key.ToString(),
	}
	res, err := c.store.Query(ctx, q)
	for e := range res.Next() {
		if e.Error != nil {
			return err
		}
		dsKey, err := core.NewDataStoreKey(e.Key)
		if err != nil {
			return err
		}

		if dsKey.InstanceType == core.ValueKey {
			err = c.store.Put(ctx, dsKey.WithDeletedFlag().ToDS(), e.Value)
			if err != nil {
				return err
			}
		}

		err = c.store.Delete(ctx, dsKey.ToDS())
		if err != nil {
			return err
		}
	}

	return nil
}

// DeltaDecode is a typed helper to extract.
// a CompositeDAGDelta from a ipld.Node
// for now let's do cbor (quick to implement)
func (c CompositeDAG) DeltaDecode(node ipld.Node) (core.Delta, error) {
	delta := &CompositeDAGDelta{}
	pbNode, ok := node.(*dag.ProtoNode)
	if !ok {
		return nil, client.NewErrUnexpectedType[*dag.ProtoNode]("ipld.Node", node)
	}
	data := pbNode.Data()
	h := &codec.CborHandle{}
	dec := codec.NewDecoderBytes(data, h)
	err := dec.Decode(delta)
	if err != nil {
		return nil, err
	}

	// get links
	for _, link := range pbNode.Links() {
		if link.Name == "head" { // ignore the head links
			continue
		}

		delta.SubDAGs = append(delta.SubDAGs, core.DAGLink{
			Name: link.Name,
			Cid:  link.Cid,
		})
	}
	return delta, nil
}
