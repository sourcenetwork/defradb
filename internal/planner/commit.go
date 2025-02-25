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
	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/encryption"
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

			if n.commitSelect.FieldID.HasValue() {
				field := n.commitSelect.FieldID.Value()
				key = key.WithFieldID(field)
			}

			n.prefix = immutable.Some[keys.HeadstoreKey](key)
		} else if n.commitSelect.FieldID.HasValue() && n.commitSelect.FieldID.Value() == "" {
			// If the user has provided an explicit nil value as `FieldID`, then we are only
			// returning collection commits.
			n.prefix = immutable.Some[keys.HeadstoreKey](keys.HeadstoreColKey{})
		}
	}

	// only need the head fetcher for non cid specific queries
	if !n.commitSelect.Cid.HasValue() {
		return n.fetcher.Start(n.planner.ctx, n.planner.txn, n.prefix, n.commitSelect.FieldID)
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
	if n.commitSelect.FieldID.HasValue() {
		fieldID = n.commitSelect.FieldID.Value()
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
	dagBlock, decBlock, err := n.loadDagBlock(currentCid)
	if err != nil {
		return false, err
	}
	currentValue, err := n.dagBlockToNodeDoc(dagBlock, decBlock)
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

func (n *dagScanNode) dagBlockToNodeDoc(dagBlock *coreblock.Block, decBlock *coreblock.Block) (core.Doc, error) {
	commit := n.commitSelect.DocumentMapping.NewDoc()
	link, err := dagBlock.GenerateLink()
	if err != nil {
		return core.Doc{}, err
	}
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.CidFieldName, link.String())

	prio := dagBlock.Delta.GetPriority()

	schemaVersionId := dagBlock.Delta.GetSchemaVersionID()
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.SchemaVersionIDFieldName, schemaVersionId)

	var fieldName any
	var fieldID any
	if dagBlock.Delta.CompositeDAGDelta != nil {
		fieldID = core.COMPOSITE_NAMESPACE
		fieldName = nil
	} else if dagBlock.Delta.CollectionDelta != nil {
		fieldID = nil
		fieldName = nil
	} else {
		fName := dagBlock.Delta.GetFieldName()
		fieldName = fName
		cols, err := n.planner.db.GetCollections(
			n.planner.ctx,
			client.CollectionFetchOptions{
				IncludeInactive: immutable.Some(true),
				SchemaVersionID: immutable.Some(schemaVersionId),
			},
		)
		if err != nil {
			return core.Doc{}, err
		}
		if len(cols) == 0 {
			return core.Doc{}, client.NewErrCollectionNotFoundForSchemaVersion(schemaVersionId)
		}

		// Because we only care about the schema, we can safely take the first - the schema is the same
		// for all in the set.
		field, ok := cols[0].Definition().GetFieldByName(fName)
		if !ok {
			return core.Doc{}, client.NewErrFieldNotExist(fName)
		}
		fieldID = field.ID.String()
	}

	if delta := dagBlock.Delta.GetData(); delta != nil {
		n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.DeltaFieldName, delta)
	} else {
		// We need to explicitely set delta to an untyped nil otherwise it will be marshalled
		// as an empty slice in the JSON response of the HTTP client.
		n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.DeltaFieldName, nil)
	}

	var deltaBytes []byte
	if !dagBlock.IsEncrypted() {
		deltaBytes = dagBlock.Delta.GetData()
	} else if decBlock != nil {
		deltaBytes = decBlock.Delta.GetData()
	}

	var deltaDecoded any
	if deltaBytes != nil {
		err := cbor.Unmarshal(deltaBytes, &deltaDecoded)
		if err != nil {
			return core.Doc{}, err
		}
	}

	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.HeightFieldName, int64(prio))
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.FieldNameFieldName, fieldName)
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.FieldIDFieldName, fieldID)
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.TypeFieldName, dagBlock.Delta.Type())
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.DeltaDecodedFieldName, deltaDecoded)

	docID := dagBlock.Delta.GetDocID()
	if docID != nil {
		n.commitSelect.DocumentMapping.SetFirstOfName(
			&commit,
			request.DocIDArgName,
			string(docID),
		)
	}

	cols, err := n.planner.db.GetCollections(
		n.planner.ctx,
		client.CollectionFetchOptions{
			IncludeInactive: immutable.Some(true),
			SchemaVersionID: immutable.Some(schemaVersionId),
		},
	)
	if err != nil {
		return core.Doc{}, err
	}
	if len(cols) == 0 {
		return core.Doc{}, client.NewErrCollectionNotFoundForSchemaVersion(schemaVersionId)
	}

	// WARNING: This will become incorrect once we allow multiple collections to share the same schema,
	// we should by then instead fetch the collection be global collection ID:
	// https://github.com/sourcenetwork/defradb/issues/1085
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit,
		request.CollectionIDFieldName, int64(cols[0].ID()))

	// links
	linksIndexes := n.commitSelect.DocumentMapping.IndexesByName[request.LinksFieldName]

	for _, linksIndex := range linksIndexes {
		links := make([]core.Doc, len(dagBlock.Heads)+len(dagBlock.Links))
		linksMapping := n.commitSelect.DocumentMapping.ChildMappings[linksIndex]

		i := 0
		for _, l := range dagBlock.Heads {
			link := linksMapping.NewDoc()
			linksMapping.SetFirstOfName(&link, request.LinksNameFieldName, "_head")
			linksMapping.SetFirstOfName(&link, request.LinksCidFieldName, l.Cid.String())

			links[i] = link
			i++
		}

		for _, l := range dagBlock.Links {
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

func (n *dagScanNode) loadDagBlock(currentCid *cid.Cid) (*coreblock.Block, *coreblock.Block, error) {
	store := n.planner.txn.Blockstore()
	block, err := store.Get(n.planner.ctx, *currentCid)
	if err != nil {
		return nil, nil, errors.Join(ErrMissingCID, err)
	}
	dagBlock, err := coreblock.GetFromBytes(block.RawData())
	if err != nil {
		return nil, nil, err
	}
	if !dagBlock.IsEncrypted() {
		return dagBlock, nil, nil
	}
	// attempt to decrypt the block if encryption data is available
	store = n.planner.txn.Encstore()
	block, err = store.Get(n.planner.ctx, dagBlock.Encryption.Cid)
	if errors.Is(err, ipld.ErrNotFound{}) {
		return dagBlock, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	encBlock, err := coreblock.GetEncryptionBlockFromBytes(block.RawData())
	if err != nil {
		return nil, nil, err
	}
	decBlock, err := encryption.DecryptBlock(n.planner.ctx, dagBlock, encBlock)
	if err != nil {
		return nil, nil, err
	}
	return dagBlock, decBlock, nil
}
