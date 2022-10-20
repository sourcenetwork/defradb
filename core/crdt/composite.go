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

	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
	"github.com/ugorji/go/codec"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
)

var (
	_ core.ReplicatedData = (*CompositeDAG)(nil)
	_ core.CompositeDelta = (*CompositeDAGDelta)(nil)
)

type CompositeDAGDelta struct {
	SchemaID string
	Priority uint64
	Data     []byte
	SubDAGs  []core.DAGLink
}

// GetPriority gets the current priority for this delta.
func (delta *CompositeDAGDelta) GetPriority() uint64 {
	return delta.Priority
}

// SetPriority will set the priority for this delta.
func (delta *CompositeDAGDelta) SetPriority(prio uint64) {
	delta.Priority = prio
}

func (delta *CompositeDAGDelta) Marshal() ([]byte, error) {
	h := &codec.CborHandle{}
	buf := bytes.NewBuffer(nil)
	enc := codec.NewEncoder(buf, h)
	err := enc.Encode(struct {
		SchemaID string
		Priority uint64
		Data     []byte
	}{delta.SchemaID, delta.Priority, delta.Data})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (delta *CompositeDAGDelta) Value() any {
	return delta.Data
}

func (delta *CompositeDAGDelta) Links() []core.DAGLink {
	return delta.SubDAGs
}

func (delta *CompositeDAGDelta) GetSchemaID() string {
	return delta.SchemaID
}

// CompositeDAG is a CRDT structure that is used
// to track a collection of sub MerkleCRDTs.
type CompositeDAG struct {
	key      string
	schemaID string
}

func NewCompositeDAG(
	store datastore.DSReaderWriter,
	schemaID string,
	namespace core.Key,
	key string,
) CompositeDAG {
	return CompositeDAG{
		key:      key,
		schemaID: schemaID,
	}
}

func (c CompositeDAG) ID() string {
	return c.key
}

func (c CompositeDAG) Value(ctx context.Context) ([]byte, error) {
	return nil, nil
}

func (c CompositeDAG) Set(patch []byte, links []core.DAGLink) *CompositeDAGDelta {
	// make sure the links are sorted lexicographically by CID
	sort.Slice(links, func(i, j int) bool {
		return strings.Compare(links[i].Cid.String(), links[j].Cid.String()) < 0
	})
	return &CompositeDAGDelta{
		Data:     patch,
		SubDAGs:  links,
		SchemaID: c.schemaID,
	}
}

// Merge implements ReplicatedData interface
// Merge two LWWRegistry based on the order of the timestamp (ts),
// if they are equal, compare IDs
// MUTATE STATE
// @todo
func (c CompositeDAG) Merge(ctx context.Context, delta core.Delta, id string) error {
	// d, ok := delta.(*CompositeDAGDelta)
	// if !ok {
	// 	return core.ErrMismatchedMergeType
	// }

	// return reg.setValue(d.Data, d.GetPriority())
	return nil
}

// DeltaDecode is a typed helper to extract
// a LWWRegDelta from a ipld.Node
// for now let's do cbor (quick to implement)
func (c CompositeDAG) DeltaDecode(node ipld.Node) (core.Delta, error) {
	delta := &CompositeDAGDelta{}
	pbNode, ok := node.(*dag.ProtoNode)
	if !ok {
		return nil, errors.New("failed to cast ipld.Node to ProtoNode")
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
