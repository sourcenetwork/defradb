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

//			   -> D1 -> E1 -> F1
// A -> B -> C |
//			   -> D2 -> E2 -> F2

/*

/db/blocks/QmKJHSDLFKJHSLDFKJHSFLDFDJKSDF => IPLD_BLOCK_BYTE_ARRAY
/db/blocks/QmJSDHGFKJSHGDKKSDGHJKFGHKSD => IPLD_BLOCK_BYTE_ARRAY
/db/blocks/QmHLSHDFLHJSDFLHJFSLDKSH => IPLD_BLOCK_BYTE_ARRAY  => []byte("hello")
/db/blocks/QmSFHLSDHLHJSDLFHJLSD => IPLD_BLOCK_BYTE_ARRA	=> []byte("zgoodbye")
/db/blocks/QmSKFJHLSDHJFLSFHD => IPLD_BLOCK_BYTE_ARRAY	=> []byte("stupid")

/db/data/1/0/bae-ALICE/1:v => "stupid"
/db/data/1/0/bae-ALICE/C:v => []byte...

/db/heads/bae-ALICE/C/QmJSDHGFKJSHGDKKSDGHJKFGHKSD => [priority=1]
/db/heads/bae-ALICE/C/QmKJHSDLFKJHSLDFKJHSFLDFDJKSDF => [priority=1]
/db/heads/bae-ALICE/1/QmSKFJHLSDHJFLSFHD => [priority=2]

*/

import (
	"container/list"
	"strings"

	"github.com/fxamacker/cbor/v2"
	blocks "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/query/graphql/mapper"
	parserTypes "github.com/sourcenetwork/defradb/query/graphql/parser/types"
)

type dagScanNode struct {
	documentIterator
	docMapper

	p     *Planner
	field string
	key   core.DataStoreKey

	// used for tracking traversal
	// note: depthLimit of 0 or 1 are equivalent
	// since the depth check is done after the
	// block scan.
	// If we need an infinite depth, use math.MaxUint32
	depthLimit   uint64
	depthVisited uint64
	visitedNodes map[string]bool

	queuedCids *list.List

	fetcher fetcher.HeadFetcher
	spans   core.Spans
	parsed  *mapper.CommitSelect
}

func (p *Planner) DAGScan(parsed *mapper.CommitSelect) *dagScanNode {
	return &dagScanNode{
		p:            p,
		visitedNodes: make(map[string]bool),
		queuedCids:   list.New(),
		parsed:       parsed,
		docMapper:    docMapper{&parsed.DocumentMapping},
	}
}

func (n *dagScanNode) Kind() string {
	return "dagScanNode"
}

func (n *dagScanNode) Init() error {
	if len(n.spans.Value) == 0 {
		n.spans = core.NewSpans(core.NewSpan(n.key, n.key.PrefixEnd()))
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

	// make sure we have the correct field suffix
	headSetSpans := core.Spans{
		HasValue: spans.HasValue,
		Value:    make([]core.Span, len(spans.Value)),
	}
	copy(headSetSpans.Value, spans.Value)
	span := headSetSpans.Value[0].Start()
	if !strings.HasSuffix(span.ToString(), n.field) {
		headSetSpans.Value[0] = core.NewSpan(span.WithFieldId(n.field), core.DataStoreKey{})
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
	if len(n.field) != 0 {
		explainerMap["field"] = n.field
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

	// find target CID either through headset or direct cid.
	if n.queuedCids.Len() > 0 {
		c := n.queuedCids.Front()
		cid, ok := c.Value.(cid.Cid)
		if !ok {
			return false, errors.New("Queued value in DAGScan isn't a CID")
		}
		n.queuedCids.Remove(c)
		currentCid = &cid
	} else if n.parsed.Cid.HasValue() && n.parsed.DocKey == "" {
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

	var heads []*ipld.Link
	n.currentValue, heads, err = n.dagBlockToNodeDoc(block)
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
	if n.depthVisited < n.depthLimit {
		// traverse the merkle dag to fetch related commits
		for _, h := range heads {
			// queue our found heads
			n.queuedCids.PushFront(h.Cid)
		}
	}

	if n.parsed.Cid.HasValue() && currentCid.String() != n.parsed.Cid.Value() {
		return n.Next()
	}

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
	n.parsed.DocumentMapping.SetFirstOfName(&commit, "cid", &cid)

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
		return core.Doc{}, nil, errors.New("Commit Delta missing priority key")
	}

	n.parsed.DocumentMapping.SetFirstOfName(&commit, "height", int64(prio))
	n.parsed.DocumentMapping.SetFirstOfName(&commit, "delta", delta["Data"])

	heads := make([]*ipld.Link, 0)

	// links
	linksIndexes := n.parsed.DocumentMapping.IndexesByName[parserTypes.LinksFieldName]

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
