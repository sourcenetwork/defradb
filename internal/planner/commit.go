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
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/keys"
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
	prefix       immutable.Option[keys.HeadstoreKey]
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
	if !n.prefix.HasValue() {
		if n.commitSelect.DocID.HasValue() {
			key := keys.HeadstoreDocKey{}.WithDocID(n.commitSelect.DocID.Value())

			if n.commitSelect.FieldName.HasValue() &&
				n.commitSelect.FieldName.Value() == request.CompositeFieldName {
				key = key.WithFieldID(core.COMPOSITE_NAMESPACE)
			}

			n.prefix = immutable.Some[keys.HeadstoreKey](key)
		} else if n.commitSelect.FieldName.HasValue() && n.commitSelect.FieldName.Value() == "" {
			// If the user has provided an explicit nil value as `FieldName`, then we are only
			// returning collection commits.
			n.prefix = immutable.Some[keys.HeadstoreKey](keys.HeadstoreColKey{})
		}
	}

	// only need the head fetcher for non cid specific queries
	if !n.commitSelect.Cid.HasValue() {
		return n.fetcher.Start(n.planner.ctx, n.planner.txn, n.prefix)
	}
	return nil
}

func (n *dagScanNode) Start() error {
	return nil
}

// Prefixes needs to parse the given prefix set. dagScanNode only
// cares about the first value in the prefix set. The value is
// either a CID or a DocID.
// If its a CID, set the node CID val
// if its a DocID, set the node Key val (headset)
func (n *dagScanNode) Prefixes(prefixes []keys.Walkable) {
	if len(prefixes) == 0 {
		return
	}

	var fieldID string
	if n.commitSelect.FieldName.HasValue() && n.commitSelect.FieldName.Value() != request.CompositeFieldName {
		// no-op, cannot use a field prefix
	} else {
		fieldID = core.COMPOSITE_NAMESPACE
	}

	for _, prefix := range prefixes {
		var start keys.HeadstoreDocKey
		switch s := prefix.(type) {
		case keys.DataStoreKey:
			start = s.ToHeadStoreKey()
		case keys.HeadstoreDocKey:
			start = s
		}

		n.prefix = immutable.Some[keys.HeadstoreKey](start.WithFieldID(fieldID))
		return
	}
}

func (n *dagScanNode) Close() error {
	if !n.commitSelect.Cid.HasValue() {
		return n.fetcher.Close()
	}
	return nil
}

func (n *dagScanNode) Source() planNode { return nil }

func (n *dagScanNode) simpleExplain() (map[string]any, error) {
	simpleExplainMap := map[string]any{}

	// Add the field attribute to the explanation if it exists.
	if n.commitSelect.FieldName.HasValue() {
		simpleExplainMap[request.FieldNameName] = n.commitSelect.FieldName.Value()
	} else {
		simpleExplainMap[request.FieldNameName] = nil
	}

	// Add the cid attribute to the explanation if it exists.
	if n.commitSelect.Cid.HasValue() {
		simpleExplainMap["cid"] = n.commitSelect.Cid.Value()
	} else {
		simpleExplainMap["cid"] = nil
	}

	// Build the explanation of the prefixes attribute.
	prefixesExplainer := []string{}
	// Note: n.headset is `nil` for single commit selection query, so must check for it.
	if n.prefix.HasValue() {
		prefixesExplainer = append(prefixesExplainer, keys.PrettyPrint(n.prefix.Value()))
	}
	// Add the built prefixes attribute, if it was valid.
	simpleExplainMap[prefixesLabel] = prefixesExplainer

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
	store := n.planner.txn.Blockstore()

	if len(n.queuedCids) > 0 {
		currentCid = n.queuedCids[0]
		n.queuedCids = n.queuedCids[1:(len(n.queuedCids))]
	} else if n.commitSelect.Cid.HasValue() && len(n.visitedNodes) == 0 {
		cid, err := cid.Parse(n.commitSelect.Cid.Value())
		if err != nil {
			return false, err
		}

		currentCid = &cid
	} else if !n.commitSelect.Cid.HasValue() {
		cid, err := n.fetcher.FetchNext()
		if err != nil || cid == nil {
			return false, err
		}

		currentCid = cid
		// Reset the depthVisited for each head yielded by headset
		n.depthVisited = 0
	} else {
		return false, nil
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
		return false, errors.Join(ErrMissingCID, err)
	}

	dagBlock, err := coreblock.GetFromBytes(block.RawData())
	if err != nil {
		return false, err
	}

	if n.commitSelect.FieldName.HasValue() {
		if n.commitSelect.FieldName.Value() == request.CompositeFieldName {
			if dagBlock.Delta.IsComposite() {
				// no-op, block passes the filter and should continue in this func
			} else {
				return n.Next()
			}
		} else {
			fieldName := dagBlock.Delta.GetFieldName()

			if fieldName == n.commitSelect.FieldName.Value() {
				// no-op, block passes the filter and should continue in this func
			} else {
				return n.Next()
			}
		}
	}

	currentValue, err := n.dagBlockToNodeDoc(dagBlock)
	if err != nil {
		return false, err
	}

	// if this is a time travel query or a latestCommits
	// (cid + undefined depth + docId) then we need to make sure the
	// target block actually belongs to the doc, since we are
	// bypassing the HeadFetcher for the first cid
	currentDocID := n.commitSelect.DocumentMapping.FirstOfName(currentValue, request.DocIDArgName)
	if n.commitSelect.Cid.HasValue() &&
		len(n.visitedNodes) == 0 &&
		n.commitSelect.DocID.HasValue() &&
		currentDocID != n.commitSelect.DocID.Value() {
		return false, ErrIncorrectCIDForDocId
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

	// the default behavior for depth is:
	// doc ID, max depth
	// just doc ID + CID, 0 depth
	// doc ID + CID + depth, use depth
	if (!n.commitSelect.Depth.HasValue() && !n.commitSelect.Cid.HasValue()) ||
		(n.commitSelect.Depth.HasValue() && n.depthVisited < n.commitSelect.Depth.Value()) {
		// Insert the newly fetched cids into the slice of queued items, in reverse order
		// so that the last new cid will be at the front of the slice
		n.queuedCids = append(make([]*cid.Cid, len(dagBlock.Heads)), n.queuedCids...)

		for i, head := range dagBlock.Heads {
			n.queuedCids[len(dagBlock.Heads)-i-1] = &head.Cid
		}
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

func (n *dagScanNode) dagBlockToNodeDoc(block *coreblock.Block) (core.Doc, error) {
	commit := n.commitSelect.DocumentMapping.NewDoc()
	link, err := block.GenerateLink()
	if err != nil {
		return core.Doc{}, err
	}
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.CidFieldName, link.String())

	schemaVersionId := block.Delta.GetSchemaVersionID()
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.SchemaVersionIDFieldName, schemaVersionId)

	cols, err := n.planner.db.GetCollections(
		n.planner.ctx,
		client.CollectionFetchOptions{
			IncludeInactive: immutable.Some(true),
			VersionID:       immutable.Some(schemaVersionId),
		},
	)
	if err != nil {
		return core.Doc{}, err
	}
	if len(cols) == 0 {
		return core.Doc{}, client.NewErrCollectionNotFoundForCollectionVersion(schemaVersionId)
	}

	var fieldName any
	if block.Delta.IsComposite() {
		fieldName = request.CompositeFieldName
	} else if block.Delta.IsCollection() {
		fieldName = nil
	} else {
		fieldName = block.Delta.GetFieldName()
	}

	// We need to explicitly set delta to an untyped nil otherwise it will be marshalled
	// as an empty slice in the JSON response of the HTTP client.
	d := block.Delta.GetData()
	if d != nil {
		n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.DeltaFieldName, d)
	} else {
		n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.DeltaFieldName, nil)
	}

	if block.Signature != nil {
		err := n.addSignatureFieldToDoc(*block.Signature, &commit)
		if err != nil {
			return core.Doc{}, err
		}
	}

	prio := block.Delta.GetPriority()
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.HeightFieldName, int64(prio))
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.FieldNameFieldName, fieldName)

	docID := block.Delta.GetDocID()
	if docID != nil {
		n.commitSelect.DocumentMapping.SetFirstOfName(
			&commit,
			request.DocIDArgName,
			string(docID),
		)
	}

	// links
	linksIndexes := n.commitSelect.DocumentMapping.IndexesByName[request.LinksFieldName]

	for _, linksIndex := range linksIndexes {
		links := make([]core.Doc, len(block.Heads)+len(block.Links))
		linksMapping := n.commitSelect.DocumentMapping.ChildMappings[linksIndex]

		i := 0
		for _, l := range block.Heads {
			link := linksMapping.NewDoc()
			linksMapping.SetFirstOfName(&link, request.LinksNameFieldName, "_head")
			linksMapping.SetFirstOfName(&link, request.LinksCidFieldName, l.Cid.String())

			links[i] = link
			i++
		}

		for _, l := range block.Links {
			link := linksMapping.NewDoc()
			if l.Name != "" {
				linksMapping.SetFirstOfName(&link, request.LinksNameFieldName, l.Name)
			}
			linksMapping.SetFirstOfName(&link, request.LinksCidFieldName, l.Link.Cid.String())

			links[i] = link
			i++
		}

		commit.Fields[linksIndex] = links
	}

	return commit, nil
}

func (n *dagScanNode) addSignatureFieldToDoc(link cidlink.Link, commit *core.Doc) error {
	store := n.planner.txn.Blockstore()
	sigIPLDBlock, err := store.Get(n.planner.ctx, link.Cid)
	if err != nil {
		return err
	}

	sigBlock, err := coreblock.GetSignatureBlockFromBytes(sigIPLDBlock.RawData())
	if err != nil {
		return err
	}
	sigFieldIndex := n.commitSelect.DocumentMapping.IndexesByName[request.SignatureFieldName][0]
	sigMapping := n.commitSelect.DocumentMapping.ChildMappings[sigFieldIndex]

	sigDoc := sigMapping.NewDoc()
	sigMapping.SetFirstOfName(&sigDoc, request.SignatureTypeFieldName, sigBlock.Header.Type)
	sigMapping.SetFirstOfName(&sigDoc, request.SignatureIdentityFieldName, sigBlock.Header.Identity)
	sigMapping.SetFirstOfName(&sigDoc, request.SignatureValueFieldName, sigBlock.Value)

	n.commitSelect.DocumentMapping.SetFirstOfName(commit, request.SignatureFieldName, sigDoc)

	return nil
}
