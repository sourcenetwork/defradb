// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"fmt"

	dag "github.com/ipfs/go-merkledag"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/core/crdt"
	"github.com/sourcenetwork/defradb/merkle/clock"
	"github.com/ugorji/go/codec"
)

// A lazily loaded cache-node that allows retrieval of cached documents at arbitrary indexes.
// The node will start empty, and then load items as they are requested.  Items that are
// requested more than once will not be re-loaded from source.
type statusScanNode struct {
	documentIterator
	docMapper

	planner *Planner
	plan    planNode
}

func (p *Planner) Status(parent *selectNode, source planNode) *statusScanNode {
	return &statusScanNode{
		planner:   p,
		plan:      source,
		docMapper: docMapper{parent.documentMapping},
	}
}

func (n *statusScanNode) Kind() string {
	return "statusScanNode"
}

func (n *statusScanNode) Init() error { return n.plan.Init() }

func (n *statusScanNode) Start() error { return n.plan.Start() }

func (n *statusScanNode) Spans(spans core.Spans) { n.plan.Spans(spans) }

func (n *statusScanNode) Close() error { return n.plan.Close() }

func (n *statusScanNode) Source() planNode { return n.plan }

func (n *statusScanNode) Value() core.Doc { return n.plan.Value() }

func (n *statusScanNode) Next() (bool, error) {
	hasNext, err := n.plan.Next()
	if err != nil || !hasNext {
		return hasNext, err
	}

	doc := n.plan.Value()

	hsKey := core.HeadStoreKey{
		DocKey:  doc.GetKey(),
		FieldId: core.COMPOSITE_NAMESPACE,
	}

	headset := clock.NewHeadSet(
		n.planner.txn.Headstore(),
		hsKey,
	)
	cids, _, err := headset.List(n.planner.ctx)
	if err != nil {
		return false, err
	}

	b, err := n.planner.txn.DAGstore().Get(n.planner.ctx, cids[0])
	if err != nil {
		return false, err
	}
	nd, err := dag.DecodeProtobuf(b.RawData())
	if err != nil {
		return false, err
	}
	delta := &crdt.CompositeDAGDelta{}
	h := &codec.CborHandle{}
	dec := codec.NewDecoderBytes(nd.Data(), h)
	err = dec.Decode(delta)
	if err != nil {
		return false, err
	}

	n.documentMapping.SetFirstOfName(&doc, request.StatusFieldName, client.DocumentStatusToString[delta.Status])
	n.currentValue = doc
	return true, nil
}

func (n *statusScanNode) Merge() bool { return true }
