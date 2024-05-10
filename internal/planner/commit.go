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
	cid "github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

type dagScanNode struct {
	documentIterator
	docMapper

	planner *Planner

	depthVisited uint64
	visitedNodes map[string]bool

	queuedCids []*cid.Cid

	fetcher      fetcher.HeadFetcher
	spans        core.Spans
	commitSelect *mapper.CommitSelect

	execInfo dagScanExecInfo
}

type dagScanExecInfo struct {
	// Total number of times dag scan was issued.
	iterations uint64
}

func (p *Planner) DAGScan(commitSelect *mapper.CommitSelect) *dagScanNode {
	return &dagScanNode{
		planner:      p,
		visitedNodes: make(map[string]bool),
		queuedCids:   []*cid.Cid{},
		commitSelect: commitSelect,
		docMapper:    docMapper{commitSelect.DocumentMapping},
	}
}

func (p *Planner) CommitSelect(commitSelect *mapper.CommitSelect) (planNode, error) {
	dagScan := p.DAGScan(commitSelect)
	return p.SelectFromSource(&commitSelect.Select, dagScan, false, nil)
}

func (n *dagScanNode) Kind() string {
	return "dagScanNode"
}

func (n *dagScanNode) Init() error {
	if len(n.spans.Value) == 0 {
		if n.commitSelect.DocID.HasValue() {
			dsKey := core.DataStoreKey{}.WithDocID(n.commitSelect.DocID.Value())

			if n.commitSelect.FieldID.HasValue() {
				field := n.commitSelect.FieldID.Value()
				dsKey = dsKey.WithFieldId(field)
			}

			n.spans = core.NewSpans(core.NewSpan(dsKey, dsKey.PrefixEnd()))
		}
	}

	return n.fetcher.Start(n.planner.ctx, n.planner.txn, n.spans, n.commitSelect.FieldID)
}

func (n *dagScanNode) Start() error {
	return nil
}

// Spans needs to parse the given span set. dagScanNode only
// cares about the first value in the span set. The value is
// either a CID or a DocID.
// If its a CID, set the node CID val
// if its a DocID, set the node Key val (headset)
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
	if n.commitSelect.FieldID.HasValue() {
		fieldId = n.commitSelect.FieldID.Value()
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

func (n *dagScanNode) simpleExplain() (map[string]any, error) {
	simpleExplainMap := map[string]any{}

	// Add the field attribute to the explanation if it exists.
	if n.commitSelect.FieldID.HasValue() {
		simpleExplainMap[request.FieldIDName] = n.commitSelect.FieldID.Value()
	} else {
		simpleExplainMap[request.FieldIDName] = nil
	}

	// Add the cid attribute to the explanation if it exists.
	if n.commitSelect.Cid.HasValue() {
		simpleExplainMap["cid"] = n.commitSelect.Cid.Value()
	} else {
		simpleExplainMap["cid"] = nil
	}

	// Build the explanation of the spans attribute.
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
	simpleExplainMap[spansLabel] = spansExplainer

	return simpleExplainMap, nil
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *dagScanNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return n.simpleExplain()

	case request.ExecuteExplain:
		return map[string]any{
			"iterations": n.execInfo.iterations,
		}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}

func (n *dagScanNode) Next() (bool, error) {
	n.execInfo.iterations++

	var currentCid *cid.Cid
	store := n.planner.txn.DAGstore()

	if len(n.queuedCids) > 0 {
		currentCid = n.queuedCids[0]
		n.queuedCids = n.queuedCids[1:(len(n.queuedCids))]
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
	block, err := store.Get(n.planner.ctx, *currentCid)
	if err != nil {
		return false, err
	}

	dagBlock, err := coreblock.GetFromBytes(block.RawData())
	if err != nil {
		return false, err
	}

	currentValue, heads, err := n.dagBlockToNodeDoc(dagBlock)
	if err != nil {
		return false, err
	}

	// the dagscan node can traverse into the merkle dag
	// based on the specified depth limit.
	// The default query operation 'latestCommit' only cares about
	// the current latest heads, so it has a depth limit
	// of 1. The query operation 'commits' doesn't have a depth
	// limit, so it will continue to traverse the graph
	// until there are no more links, and no more explored
	// HEAD paths.
	n.depthVisited++
	n.visitedNodes[currentCid.String()] = true // mark the current node as "visited"
	if !n.commitSelect.Depth.HasValue() || n.depthVisited < n.commitSelect.Depth.Value() {
		// Insert the newly fetched cids into the slice of queued items, in reverse order
		// so that the last new cid will be at the front of the slice
		n.queuedCids = append(make([]*cid.Cid, len(heads)), n.queuedCids...)

		for i, h := range heads {
			link := h // TODO remove when Go 1.22 #2431
			n.queuedCids[len(heads)-i-1] = &link.Cid
		}
	}

	if n.commitSelect.Cid.HasValue() && currentCid.String() != n.commitSelect.Cid.Value() {
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
dagScanNode is the request plan graph node responsible for scanning through the dag
blocks of the MerkleCRDTs.

The current available endpoints are:
 - latestCommit: Given a docid, and optionally a field name, return the latest dag commit
 - commits: Given a docid, and optionally a field name, return all the dag commits

Additionally, theres a subselection available on the Document query called _version,
which returns the current dag commit for the stored CRDT value.

All the dagScanNode endpoints use similar structures
*/

func (n *dagScanNode) dagBlockToNodeDoc(block *coreblock.Block) (core.Doc, []cidlink.Link, error) {
	commit := n.commitSelect.DocumentMapping.NewDoc()
	link, err := block.GenerateLink()
	if err != nil {
		return core.Doc{}, nil, err
	}
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.CidFieldName, link.String())

	prio := block.Delta.GetPriority()

	schemaVersionId := block.Delta.GetSchemaVersionID()
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.SchemaVersionIDFieldName, schemaVersionId)

	var fieldName any

	var fieldID string
	if block.Delta.CompositeDAGDelta != nil {
		fieldID = core.COMPOSITE_NAMESPACE
		fieldName = nil
	} else {
		fName := block.Delta.GetFieldName()
		fieldName = fName
		cols, err := n.planner.db.GetCollections(
			n.planner.ctx,
			client.CollectionFetchOptions{
				IncludeInactive: immutable.Some(true),
				SchemaVersionID: immutable.Some(schemaVersionId),
			},
		)
		if err != nil {
			return core.Doc{}, nil, err
		}
		if len(cols) == 0 {
			return core.Doc{}, nil, client.NewErrCollectionNotFoundForSchemaVersion(schemaVersionId)
		}

		// Because we only care about the schema, we can safely take the first - the schema is the same
		// for all in the set.
		field, ok := cols[0].Definition().GetFieldByName(fName)
		if !ok {
			return core.Doc{}, nil, client.NewErrFieldNotExist(fName)
		}
		fieldID = field.ID.String()
	}
	// We need to explicitely set delta to nil otherwise it will be marshalled
	// as an empty slice in the JSON response of the HTTP client.
	d := block.Delta.GetData()
	if d != nil {
		n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.DeltaFieldName, d)
	} else {
		n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.DeltaFieldName, nil)
	}
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.HeightFieldName, int64(prio))
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.FieldNameFieldName, fieldName)
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.FieldIDFieldName, fieldID)

	docID := block.Delta.GetDocID()

	n.commitSelect.DocumentMapping.SetFirstOfName(&commit,
		request.DocIDArgName, string(docID))

	cols, err := n.planner.db.GetCollections(
		n.planner.ctx,
		client.CollectionFetchOptions{
			IncludeInactive: immutable.Some(true),
			SchemaVersionID: immutable.Some(schemaVersionId),
		},
	)
	if err != nil {
		return core.Doc{}, nil, err
	}
	if len(cols) == 0 {
		return core.Doc{}, nil, client.NewErrCollectionNotFoundForSchemaVersion(schemaVersionId)
	}

	// WARNING: This will become incorrect once we allow multiple collections to share the same schema,
	// we should by then instead fetch the collection be global collection ID:
	// https://github.com/sourcenetwork/defradb/issues/1085
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit,
		request.CollectionIDFieldName, int64(cols[0].ID()))

	heads := make([]cidlink.Link, 0)

	// links
	linksIndexes := n.commitSelect.DocumentMapping.IndexesByName[request.LinksFieldName]

	for _, linksIndex := range linksIndexes {
		links := make([]core.Doc, len(block.Links))
		linksMapping := n.commitSelect.DocumentMapping.ChildMappings[linksIndex]

		for i, l := range block.Links {
			link := linksMapping.NewDoc()
			linksMapping.SetFirstOfName(&link, request.LinksNameFieldName, l.Name)
			linksMapping.SetFirstOfName(&link, request.LinksCidFieldName, l.Link.Cid.String())

			links[i] = link
		}

		commit.Fields[linksIndex] = links
	}

	for _, l := range block.Links {
		if l.Name == "_head" {
			heads = append(heads, l.Link)
		}
	}

	return commit, heads, nil
}
