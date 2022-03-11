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
	"context"
	"fmt"

	"github.com/sourcenetwork/defradb/core"
	corenet "github.com/sourcenetwork/defradb/core/net"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/sourcenetwork/defradb/logging"
)

var (
	log = logging.MustNewLogger("defra.merklecrdt")
)

// MerkleCRDT is the implementation of a Merkle Clock along with a
// CRDT payload. It implements the ReplicatedData interface
// so it can be merged with any given semantics.
type MerkleCRDT interface {
	core.ReplicatedData
	Clock() core.MerkleClock
}

// type MerkleCRDTInitFn func(core.Key) MerkleCRDT
// type MerkleCRDTFactory func(store datastore.DSReaderWriter, namespace core.Key) MerkleCRDTInitFn

// Type indicates MerkleCRDT type
// type Type byte

// const (
// 	//no lint
// 	none = Type(iota) // reserved none type
// 	LWW_REGISTER
// 	OBJECT
// )

var (
	// defaultMerkleCRDTs                     = make(map[Type]MerkleCRDTFactory)
	_ core.ReplicatedData = (*baseMerkleCRDT)(nil)
)

// The baseMerkleCRDT handles the merkle crdt overhead functions
// that aren't CRDT specific like the mutations and state retrieval
// functions. It handles creating and publishing the crdt DAG with
// the help of the MerkleClock
type baseMerkleCRDT struct {
	clock core.MerkleClock
	crdt  core.ReplicatedData

	broadcaster corenet.Broadcaster
}

func (base *baseMerkleCRDT) Clock() core.MerkleClock {
	return base.clock
}

func (base *baseMerkleCRDT) Merge(ctx context.Context, other core.Delta, id string) error {
	return base.crdt.Merge(ctx, other, id)
}

func (base *baseMerkleCRDT) DeltaDecode(node ipld.Node) (core.Delta, error) {
	return base.crdt.DeltaDecode(node)
}

func (base *baseMerkleCRDT) Value(ctx context.Context) ([]byte, error) {
	return base.crdt.Value(ctx)
}

func (base *baseMerkleCRDT) ID() string {
	return base.crdt.ID()
}

// Publishes the delta to state
func (base *baseMerkleCRDT) Publish(ctx context.Context, delta core.Delta) (cid.Cid, ipld.Node, error) {
	log.Debug(ctx, "Processing CRDT state", logging.NewKV("DocKey", base.crdt.ID()))
	c, nd, err := base.clock.AddDAGNode(ctx, delta)
	if err != nil {
		return cid.Undef, nil, err
	}
	return c, nd, nil
}

func (base *baseMerkleCRDT) Broadcast(ctx context.Context, nd ipld.Node, delta core.Delta) error {
	if base.broadcaster == nil {
		return nil // just skip if we dont have a broadcaster set
	}

	parts := ds.NewKey(base.crdt.ID()).List()
	if len(parts) < 3 {
		return fmt.Errorf("Invalid dockey for MerkleCRDT")
	}
	dockey := parts[2]

	c := nd.Cid()
	netdelta, ok := delta.(core.NetDelta)
	if !ok {
		return fmt.Errorf("Can't broadcast a delta payload that doesn't implement core.NetDelta")
	}

	log.Debug(ctx, "Broadcasting new DAG node", logging.NewKV("DocKey", dockey), logging.NewKV("Cid", c))
	// we dont want to wait around for the broadcast
	go func() {
		lg := core.Log{
			DocKey:   dockey,
			Cid:      c,
			SchemaID: netdelta.GetSchemaID(),
			Block:    nd,
			Priority: netdelta.GetPriority(),
		}
		if err := base.broadcaster.Send(lg); err != nil {
			log.ErrorE(ctx, "Failed to broadcast MerkleCRDT update", err, logging.NewKV("DocKey", dockey), logging.NewKV("Cid", c))
		}
	}()

	return nil
}
