// Copyright 2020 Source Inc.
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
	"container/list"
	"strings"

	// "errors"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/fetcher"

	"github.com/fxamacker/cbor/v2"
	blocks "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
	"github.com/pkg/errors"
)

type headsetScanNode struct {
	p   *Planner
	key core.DataStoreKey

	spans           core.Spans
	scanInitialized bool

	cid *cid.Cid

	fetcher fetcher.HeadFetcher
}

func (h *headsetScanNode) Init() error {
	return h.initScan()
}

func (h *headsetScanNode) Spans(spans core.Spans) {
	h.spans = spans
}

func (h *headsetScanNode) Start() error {
	return nil
}

func (h *headsetScanNode) initScan() error {
	if len(h.spans) == 0 {
		h.spans = append(h.spans, core.NewSpan(h.key, h.key.PrefixEnd()))
	}

	// fmt.Println("startin fetcher with spans:", h.spans[0].Start())
	err := h.fetcher.Start(h.p.txn, h.spans)
	if err != nil {
		return err
	}

	h.scanInitialized = true
	return nil
}

func (h *headsetScanNode) Next() (bool, error) {
	if !h.scanInitialized {
		if err := h.initScan(); err != nil {
			return false, err
		}
	}

	var err error
	h.cid, err = h.fetcher.FetchNext()
	if err != nil {
		return false, err
	}
	if h.cid == nil {
		return false, nil
	}
	return true, nil
}

func (h *headsetScanNode) Values() map[string]interface{} {
	return map[string]interface{}{
		"cid": *h.cid,
	}
}

func (h *headsetScanNode) Close() {}

func (h *headsetScanNode) Source() planNode { return nil }

func (p *Planner) HeadScan() *headsetScanNode {
	return &headsetScanNode{p: p}
}

type dagScanNode struct {
	p *Planner

	key   *core.Key
	cid   *cid.Cid
	field string

	// used for tracking traversal
	// note: depthLimit of 0 or 1 are equivalent
	// since the depth check is done after the
	// block scan.
	// If we need an infinite depth, use math.MaxUint32
	depthLimit   uint32
	depthVisited uint32
	visitedNodes map[string]bool

	queuedCids *list.List

	headset *headsetScanNode

	// previousScanNode planNode
	// linksScanNode    planNode

	// block blocks.Block
	doc map[string]interface{}
}

func (p *Planner) DAGScan() *dagScanNode {
	return &dagScanNode{
		p:            p,
		visitedNodes: make(map[string]bool),
		queuedCids:   list.New(),
	}
}

func (n *dagScanNode) Init() error {
	if n.headset != nil {
		return n.headset.Init()
	}
	return nil
}
func (n *dagScanNode) Start() error {
	if n.headset != nil {
		return n.headset.Start()
	}
	return nil
}

// Spans needs to parse the given span set. dagScanNode only
// cares about the first value in the span set. The value is
// either a CID or a DocKey.
// If its a CID, set the node CID val
// if its a DocKey, set the node Key val (headset)
func (n *dagScanNode) Spans(spans core.Spans) {
	if len(spans) == 0 {
		return
	}

	// if we have a headset, pass along
	// otherwise, try to parse as a CID
	if n.headset != nil {
		// make sure we have the correct field suffix
		span := spans[0].Start()
		if !strings.HasSuffix(span.ToString(), n.field) {
			spans[0] = core.NewSpan(span.WithFieldId(n.field), core.DataStoreKey{})
		}
		n.headset.Spans(spans)
	} else {
		data := spans[0].Start().ToString()
		c, err := cid.Decode(data)
		if err == nil {
			n.cid = &c
		}
	}
}

func (n *dagScanNode) Close() {
	if n.headset != nil {
		n.headset.Close()
	}
}

func (n *dagScanNode) Source() planNode { return n.headset }

func (n *dagScanNode) Next() (bool, error) {
	// find target cid either through headset or direct cid.
	// if n.cid == nil {

	if n.queuedCids.Len() > 0 {
		c := n.queuedCids.Front()
		cid, ok := c.Value.(cid.Cid)
		if !ok {
			return false, errors.New("Queued value in DAGScan isn't a CID")
		}
		n.queuedCids.Remove(c)
		n.cid = &cid
	} else if n.headset != nil {
		if next, err := n.headset.Next(); !next {
			return false, err
		}

		val := n.headset.Values()
		cid, ok := val["cid"].(cid.Cid)
		if !ok {
			return false, errors.New("Headset scan node returned an invalid cid")
		}
		n.cid = &cid

	} else if n.cid == nil {
		// add this final elseif case in case another function
		// manually sets the CID. Should prob migrate any remote CID
		// updates to use the queuedCids.
		return false, nil // no queued cids and no headset available
	}

	// skip already visited CIDs
	// we only need to call Next() again
	// as it will reset and scan through the headset/queue
	// and eventually return a value, or false if we've
	// visited everything
	if _, ok := n.visitedNodes[n.cid.String()]; ok {
		n.cid = nil
		return n.Next()
	}

	// use the stored cid to scan through the blockstore
	// clear the cid after
	store := n.p.txn.DAGstore()
	block, err := store.Get(*n.cid)
	if err != nil { // handle error?
		return false, err
	}
	var heads []*ipld.Link
	n.doc, heads, err = dagBlockToNodeMap(block)
	if err != nil {
		return false, err
	}

	// the dagscan node can traverse into the merkle dag
	// based on the specified depth limit.
	// The default query 'latestCommit' only cares about
	// the current latest heads, so it has a depth limit
	// of 1. The query 'allCommits' doesn't have a depth
	// limit, so it will continue to traverse the graph
	// until there are no more links, and no more explored
	// HEAD paths.
	n.depthVisited++
	n.visitedNodes[n.cid.String()] = true // mark the current node as "visited"
	if n.depthVisited < n.depthLimit {
		// look for HEAD links
		for _, h := range heads {
			// queue our found heads
			n.queuedCids.PushFront(h.Cid)
		}

	}
	n.cid = nil // clear cid for next round
	return true, nil
}

// func (n *dagScanNode) nextHead() (cid.Cid, error) {

// }

func (n *dagScanNode) Values() map[string]interface{} {
	return n.doc
}

/*
dagScanNode is the query plan graph node responsible for scanning through the dag
blocks of the MerkleCRDTs.

The current available endpoints are:
 - latestCommit: Given a docid, and optionally a field name, return the latest dag commit
 - allCommits: Given a docid, and optionally a field name, return all the dag commits
 - oneCommit: Given a cid, return the specific commit

Additionally, theres a subselection available on the Document query called _version,
which returns the current dag commit for the stored CRDT value.

All the dagScanNode endpoints use similar structures
*/

func dagBlockToNodeMap(block blocks.Block) (map[string]interface{}, []*ipld.Link, error) {
	commit := map[string]interface{}{
		"cid": block.Cid().String(),
	}

	// decode the delta, get the priority and payload
	nd, err := dag.DecodeProtobuf(block.RawData())
	if err != nil {
		return nil, nil, err
	}

	// @todo: Wrap delta unmarshaling into a proper typed interface.
	var delta map[string]interface{}
	if err := cbor.Unmarshal(nd.Data(), &delta); err != nil {
		return nil, nil, err
	}

	prio, ok := delta["Priority"].(uint64)
	if !ok {
		return nil, nil, errors.New("Commit Delta missing priority key")
	}

	commit["height"] = int64(prio)
	commit["delta"] = delta["Data"] // check

	heads := make([]*ipld.Link, 0)

	// links
	links := make([]map[string]interface{}, len(nd.Links()))
	for i, l := range nd.Links() {
		link := map[string]interface{}{
			"name": l.Name,
			"cid":  l.Cid.String(),
		}
		links[i] = link

		if l.Name == "_head" {
			heads = append(heads, l)
		}
	}
	commit["links"] = links
	return commit, heads, nil
}
