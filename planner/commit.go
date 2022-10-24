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
	"github.com/fxamacker/cbor/v2"
	blocks "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/mapper"
)

type dagScanNode struct {
	documentIterator
	docMapper

	p *Planner

	depthVisited uint64
	visitedNodes map[string]bool

	queuedCids []*cid.Cid

	fetcher fetcher.HeadFetcher
	spans   core.Spans
	parsed  *mapper.CommitSelect
}

func (p *Planner) DAGScan(parsed *mapper.CommitSelect) *dagScanNode {
	return &dagScanNode{
		p:            p,
		visitedNodes: make(map[string]bool),
		queuedCids:   []*cid.Cid{},
		parsed:       parsed,
		docMapper:    docMapper{&parsed.DocumentMapping},
	}
}

func (p *Planner) CommitSelect(parsed *mapper.CommitSelect) (planNode, error) {
	dagScan := p.DAGScan(parsed)
	return p.SelectFromSource(&parsed.Select, dagScan, false, nil)
}

func (n *dagScanNode) Kind() string {
	return "dagScanNode"
}

func (n *dagScanNode) Init() error {
	if len(n.spans.Value) == 0 {
		if n.parsed.DocKey.HasValue() {
			key := core.DataStoreKey{}.WithDocKey(n.parsed.DocKey.Value())

			if n.parsed.FieldName.HasValue() {
				field := n.parsed.FieldName.Value()
				key = key.WithFieldId(field)
			}

			n.spans = core.NewSpans(core.NewSpan(key, key.PrefixEnd()))
		}
	}

	return n.fetcher.Start(n.p.ctx, n.p.txn, n.spans, n.parsed.FieldName)
}

func (n *dagScanNode) Start() error {
	return nil
}

// Spans needs to parse the given span set. dagScanNode only
// cares about the first value in the span set. The value is
// either a CID or a DocKey.
// If its a CID, set the node CID val
// if its a DocKey, set the node Key val (headset)
func (n *dagScanNode) Spans(spans core.Spans) {
	if len(spans.Value) == 0 {
		return
	}

	// copy the input spans so that we may mutate freely
	headSetSpans := core.Spans{
		HasValue: spans.HasValue,
		Value:    make([]core.Span, len(spans.Value)),
	}
	copy(headSetSpans.Value, spans.Value)

	var fieldId string
	if n.parsed.FieldName.HasValue() {
		fieldId = n.parsed.FieldName.Value()
	} else {
		fieldId = core.COMPOSITE_NAMESPACE
	}

	for i, span := range headSetSpans.Value {
		if span.Start().FieldId != fieldId {
			headSetSpans.Value[i] = core.NewSpan(span.Start().WithFieldId(fieldId), core.DataStoreKey{})
		}
	}

	n.spans = headSetSpans
}

func (n *dagScanNode) Close() error {
	return n.fetcher.Close()
}

func (n *dagScanNode) Source() planNode { return nil }

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *dagScanNode) Explain() (map[string]any, error) {
	explainerMap := map[string]any{}

	// Add the field attribute to the explaination if it exists.
	if n.parsed.FieldName.HasValue() {
		explainerMap["field"] = n.parsed.FieldName.Value()
	} else {
		explainerMap["field"] = nil
	}

	// Add the cid attribute to the explaination if it exists.
	if n.parsed.Cid.HasValue() {
		explainerMap["cid"] = n.parsed.Cid.Value()
	} else {
		explainerMap["cid"] = nil
	}

	// Build the explaination of the spans attribute.
	spansExplainer := []map[string]any{}
	// Note: n.headset is `nil` for single commit selection query, so must check for it.
	if n.spans.HasValue {
		for _, span := range n.spans.Value {
			spansExplainer = append(
				spansExplainer,
				map[string]any{
					"start": span.Start().ToString(),
					"end":   span.End().ToString(),
				},
			)
		}
	}
	// Add the built spans attribute, if it was valid.
	explainerMap[spansLabel] = spansExplainer

	return explainerMap, nil
}

func (n *dagScanNode) Next() (bool, error) {
	var currentCid *cid.Cid
	store := n.p.txn.DAGstore()

	if len(n.queuedCids) > 0 {
		currentCid = n.queuedCids[0]
		n.queuedCids = n.queuedCids[1:(len(n.queuedCids))]
	} else if n.parsed.Cid.HasValue() && !n.parsed.DocKey.HasValue() {
		if n.visitedNodes[n.parsed.Cid.Value()] {
			// If the requested cid has been visited, we are done and should return false
			return false, nil
		}

		cid, err := cid.Decode(n.parsed.Cid.Value())
		if err != nil {
			return false, err
		}

		if hasCid, err := store.Has(n.p.ctx, cid); !hasCid || err != nil {
			return false, err
		}

		currentCid = &cid
	} else {
		cid, err := n.fetcher.FetchNext()
		if err != nil || cid == nil {
			return false, err
		}

		currentCid = cid
		// Reset the depthVisited for each head yielded by headset
		n.depthVisited = 0
	}

	// skip already visited CIDs
	// we only need to call Next() again
	// as it will reset and scan through the headset/queue
	// and eventually return a value, or false if we've
	// visited everything
	if _, ok := n.visitedNodes[currentCid.String()]; ok {
		return n.Next()
	}

	// use the stored cid to scan through the blockstore
	// clear the cid after
	block, err := store.Get(n.p.ctx, *currentCid)
	if err != nil {
		return false, err
	}

	currentValue, heads, err := n.dagBlockToNodeDoc(block)
	if err != nil {
		return false, err
	}

	// the dagscan node can traverse into the merkle dag
	// based on the specified depth limit.
	// The default query 'latestCommit' only cares about
	// the current latest heads, so it has a depth limit
	// of 1. The query 'commits' doesn't have a depth
	// limit, so it will continue to traverse the graph
	// until there are no more links, and no more explored
	// HEAD paths.
	n.depthVisited++
	n.visitedNodes[currentCid.String()] = true // mark the current node as "visited"
	if !n.parsed.Depth.HasValue() || n.depthVisited < n.parsed.Depth.Value() {
		// Insert the newly fetched cids into the slice of queued items, in reverse order
		// so that the last new cid will be at the front of the slice
		n.queuedCids = append(make([]*cid.Cid, len(heads)), n.queuedCids...)

		for i, h := range heads {
			n.queuedCids[len(heads)-i-1] = &h.Cid
		}
	}

	if n.parsed.Cid.HasValue() && currentCid.String() != n.parsed.Cid.Value() {
		// If a specific cid has been requested, and the current item does not
		// match, keep searching.
		return n.Next()
	}

	n.currentValue = currentValue
	return true, nil
}

//			   -> D1 -> E1 -> F1
// A -> B -> C |
//			   -> D2 -> E2 -> F2

/*

/db/blocks/QmKJHSDLFKJHSLDFKJHSFLDFDJKSDF => IPLD_BLOCK_BYTE_ARRAY
/db/blocks/QmJSDHGFKJSHGDKKSDGHJKFGHKSD => IPLD_BLOCK_BYTE_ARRAY
/db/blocks/QmHLSHDFLHJSDFLHJFSLDKSH => IPLD_BLOCK_BYTE_ARRAY  => []byte("hello")
/db/blocks/QmSFHLSDHLHJSDLFHJLSD => IPLD_BLOCK_BYTE_ARRAY	=> []byte("goodbye")
/db/data/1/0/bae-ALICE/1:v => "hello"
/db/data/1/0/bae-ALICE/C:v => []byte...
/db/heads/bae-ALICE/C/QmJSDHGFKJSHGDKKSDGHJKFGHKSD => [priority=1]
/db/heads/bae-ALICE/C/QmKJHSDLFKJHSLDFKJHSFLDFDJKSDF => [priority=1]
/db/heads/bae-ALICE/1/QmHLSHDFLHJSDFLHJFSLDKSH => [priority=1]
/db/heads/bae-ALICE/1/QmSFHLSDHLHJSDLFHJLSD => [priority=1]

*/

// func (n *dagScanNode) nextHead() (cid.Cid, error) {

// }

/*
dagScanNode is the query plan graph node responsible for scanning through the dag
blocks of the MerkleCRDTs.

The current available endpoints are:
 - latestCommit: Given a docid, and optionally a field name, return the latest dag commit
 - commits: Given a docid, and optionally a field name, return all the dag commits

Additionally, theres a subselection available on the Document query called _version,
which returns the current dag commit for the stored CRDT value.

All the dagScanNode endpoints use similar structures
*/

func (n *dagScanNode) dagBlockToNodeDoc(block blocks.Block) (core.Doc, []*ipld.Link, error) {
	commit := n.parsed.DocumentMapping.NewDoc()
	cid := block.Cid()
	n.parsed.DocumentMapping.SetFirstOfName(&commit, "cid", cid.String())

	// decode the delta, get the priority and payload
	nd, err := dag.DecodeProtobuf(block.RawData())
	if err != nil {
		return core.Doc{}, nil, err
	}

	// @todo: Wrap delta unmarshaling into a proper typed interface.
	var delta map[string]any
	if err := cbor.Unmarshal(nd.Data(), &delta); err != nil {
		return core.Doc{}, nil, err
	}

	prio, ok := delta["Priority"].(uint64)
	if !ok {
		return core.Doc{}, nil, errors.New("commit Delta missing priority key")
	}

	n.parsed.DocumentMapping.SetFirstOfName(&commit, "height", int64(prio))
	n.parsed.DocumentMapping.SetFirstOfName(&commit, "delta", delta["Data"])

	heads := make([]*ipld.Link, 0)

	// links
	linksIndexes := n.parsed.DocumentMapping.IndexesByName[request.LinksFieldName]

	for _, linksIndex := range linksIndexes {
		links := make([]core.Doc, len(nd.Links()))
		linksMapping := n.parsed.DocumentMapping.ChildMappings[linksIndex]

		for i, l := range nd.Links() {
			link := linksMapping.NewDoc()
			linksMapping.SetFirstOfName(&link, "name", l.Name)
			linksMapping.SetFirstOfName(&link, "cid", l.Cid.String())

			links[i] = link
		}

		commit.Fields[linksIndex] = links
	}

	for _, l := range nd.Links() {
		if l.Name == "_head" {
			heads = append(heads, l)
		}
	}

	return commit, heads, nil
}

func (n *dagScanNode) Append() bool { return true }
