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
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/core"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/datastore"
	acpDB "github.com/sourcenetwork/defradb/internal/db/acp"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

type dagScanNode struct {
	documentIterator
	docMapper

	planner *Planner

	depthVisited uint64
	visitedNodes map[string]bool

	queuedCids            []*cid.Cid
	currentSelectCidIndex int

	// queuedCommits holds commits produced from the current block that have not yet been
	// yielded. A field block shared by several documents (identical genesis field deltas)
	// produces one commit per owning document when queried by CID without a docID.
	queuedCommits []core.Doc

	fetcher        fetcher.HeadFetcher
	fetcherStarted bool
	noResults      bool
	prefix         immutable.Option[keys.HeadstoreKey]
	commitSelect   *mapper.CommitSelect

	linksScanNodes []*dagScanNode
	headsScanNodes []*dagScanNode

	activeDocID      immutable.Option[string]
	activeDocShortID immutable.Option[uint64]

	execInfo dagScanExecInfo
}

type dagScanExecInfo struct {
	// Total number of times dag scan was issued.
	iterations uint64
}

func (p *Planner) DAGScan(commitSelect *mapper.CommitSelect) *dagScanNode {
	node := &dagScanNode{
		planner:      p,
		visitedNodes: make(map[string]bool),
		queuedCids:   []*cid.Cid{},
		commitSelect: commitSelect,
		docMapper:    docMapper{commitSelect.DocumentMapping},
	}

	//  Add the sub dagScan planNodes to handle the
	// "links" commit selection
	for _, f := range commitSelect.Fields {
		switch innerCommit := f.(type) {
		case *mapper.CommitSelect:
			// links only go a max depth of one. If you want to
			// go deeper, use nested "links" fields
			innerCommit.Depth = immutable.Some(uint64(0))
			innerNode := p.DAGScan(innerCommit)

			switch innerCommit.Field.Name {
			case request.LinksFieldName:
				node.linksScanNodes = append(node.linksScanNodes, innerNode)
			case request.HeadsFieldName:
				node.headsScanNodes = append(node.headsScanNodes, innerNode)
			}
		}
	}

	return node
}

func (p *Planner) CommitSelect(commitSelect *mapper.CommitSelect) (planNode, error) {
	dagScan := p.DAGScan(commitSelect)
	return p.SelectFromSource(&commitSelect.Select, dagScan, false, nil)
}

func (n *dagScanNode) Kind() string {
	return "dagScanNode"
}

func (n *dagScanNode) Init() error {
	if !n.prefix.HasValue() && !n.commitSelect.Cids.HasValue() {
		if n.commitSelect.DocIDs.HasValue() && len(n.commitSelect.DocIDs.Value()) > 0 {
			docRef, found, err := n.getHeadstoreDocRef(n.commitSelect.DocIDs.Value()[0])
			if err != nil {
				return err
			}
			if !found {
				n.noResults = true
				return nil
			}
			key := keys.HeadstoreDocKey{}.WithDocShortID(docRef.DocShortID)
			n.prefix = immutable.Some[keys.HeadstoreKey](key)
		}
	}

	if n.noResults {
		return nil
	}

	// only need the head fetcher for non cid specific queries
	if !n.commitSelect.Cids.HasValue() && len(n.queuedCids) == 0 {
		n.fetcherStarted = true
		return n.fetcher.Start(n.planner.ctx, n.prefix)
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

	var start keys.HeadstoreDocKey
	switch s := prefixes[0].(type) {
	case keys.DataStoreKey:
		start = s.ToHeadStoreKey()
	case keys.HeadstoreDocKey:
		start = s
	}

	n.prefix = immutable.Some[keys.HeadstoreKey](start.WithFieldID(core.COMPOSITE_NAMESPACE))
}

func (n *dagScanNode) getHeadstoreDocRef(docID string) (keys.DocRef, bool, error) {
	if docID == "" {
		return keys.DocRef{}, false, nil
	}
	docRef, found, err := id.GetDocRef(n.planner.ctx, docID)
	if err != nil {
		return keys.DocRef{}, false, err
	}
	return docRef, found, nil
}

func (n *dagScanNode) Close() error {
	if !n.commitSelect.Cids.HasValue() {
		return n.fetcher.Close()
	}
	return nil
}

func (n *dagScanNode) Source() planNode { return nil }

func (n *dagScanNode) simpleExplain() (map[string]any, error) {
	simpleExplainMap := map[string]any{}

	// Add the cid attribute to the explanation if it exists.
	if n.commitSelect.Cids.HasValue() {
		simpleExplainMap["cid"] = n.commitSelect.Cids.Value()
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
	txn := datastore.CtxMustGetTxn(n.planner.ctx)

	n.execInfo.iterations++

	if n.noResults {
		return false, nil
	}

	// Yield any commits already produced from the current block (a shared field block
	// fanned out across its owning documents) before advancing to the next block.
	if len(n.queuedCommits) > 0 {
		n.currentValue = n.queuedCommits[0]
		n.queuedCommits = n.queuedCommits[1:]
		return true, nil
	}

	var currentCid *cid.Cid

	if len(n.queuedCids) > 0 {
		currentCid = n.queuedCids[0]
		n.queuedCids = n.queuedCids[1:(len(n.queuedCids))]
	} else if n.commitSelect.Cids.HasValue() && n.currentSelectCidIndex < len(n.commitSelect.Cids.Value()) {
		cid, err := cid.Parse(n.commitSelect.Cids.Value()[n.currentSelectCidIndex])
		if err != nil {
			return false, err
		}

		currentCid = &cid
		n.currentSelectCidIndex++
	} else if !n.commitSelect.Cids.HasValue() && n.fetcherStarted {
		cid, err := n.fetcher.FetchNext()
		if err != nil || cid == nil {
			return false, err
		}

		currentCid = cid
		n.setActiveDocIDFromCurrentHead()
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
	block, err := txn.Blockstore().Get(n.planner.ctx, *currentCid)
	if err != nil {
		return false, errors.Join(ErrIncorrectOrMissingCID, err)
	}

	dagBlock, err := coreblock.GetFromBytes(block.RawData())
	if err != nil {
		return false, err
	}

	// Resolve the documents this block's commit should be reported under. A field block
	// shared across documents yields several owners; a docID-scoped traversal yields one.
	docIDs, err := n.commitDocIDs(dagBlock, *currentCid)
	if err != nil {
		return false, err
	}

	// We want to enforce DAC, so we skip commits the caller has no read access to.
	// We do this before [dagBlockToNodeDoc], so denied commits do not pay the
	// doc-mapping cost. We only check when document acp is configured.
	//
	// A commit is gated by the read access of the document(s) it belongs to: an explicit grant on a
	// document is sufficient on its own, otherwise — for a branchable collection — the caller also
	// needs access to the collection object, so a private branchable collection gates its entire
	// commit DAG. A shared field block has several owners; the commit is kept for those the caller
	// may read. A collection-level commit carries no docID and is gated on the collection object for
	// a branchable collection (and is not gated otherwise, e.g. schema-definition changes).
	if n.planner.documentACP.HasValue() {
		versionID := dagBlock.Delta.GetCollectionVersionID()

		// Pass the requester's identity so the collection lookup is authorised as them (rather than
		// anonymously) when node acp gates the get-collection operation.
		getColOpts := options.GetCollections().SetGetInactive(true).SetVersionID(versionID)
		if n.planner.identity.HasValue() {
			getColOpts = getColOpts.SetIdentity(n.planner.identity.Value())
		}
		cols, err := n.planner.db.GetCollections(n.planner.ctx, getColOpts)
		if err != nil {
			return false, err
		}
		if len(cols) == 0 {
			return false, client.NewErrCollectionNotFoundForCollectionVersion(versionID)
		}

		if dagBlock.Delta.IsCollection() {
			// Collection-level commit: no docID. CheckDocReadAccess gates it on the collection
			// object for a branchable collection, and is a no-op otherwise.
			hasPermission, err := acpDB.CheckDocReadAccess(
				n.planner.ctx,
				n.planner.identity,
				n.planner.nodeACP,
				n.planner.documentACP.Value(),
				cols[0],
				"",
			)
			if err != nil {
				return false, err
			}
			if !hasPermission {
				n.visitedNodes[currentCid.String()] = true
				return n.Next()
			}
		} else {
			if len(docIDs) == 0 {
				return false, client.ErrMalformedDocID
			}

			// A shared field block has several owners; keep only those the caller may read.
			permitted := make([]string, 0, len(docIDs))
			for _, docID := range docIDs {
				hasPermission, err := acpDB.CheckDocReadAccess(
					n.planner.ctx,
					n.planner.identity,
					n.planner.nodeACP,
					n.planner.documentACP.Value(),
					cols[0],
					docID,
				)
				if err != nil {
					return false, err
				}
				if hasPermission {
					permitted = append(permitted, docID)
				}
			}
			if len(permitted) == 0 {
				// Mark visited so the same denied cid is not re-checked if it
				// reappears via a links/heads queue elsewhere in the traversal.
				n.visitedNodes[currentCid.String()] = true
				return n.Next()
			}
			docIDs = permitted
		}
	}

	// Build one commit per resolved document. A block with no associated document (e.g. a
	// collection-level delta) still yields a single commit with no docID.
	commits := make([]core.Doc, 0, max(len(docIDs), 1))
	if len(docIDs) == 0 {
		commit, err := n.dagBlockToNodeDoc(dagBlock, "")
		if err != nil {
			return false, err
		}
		commits = append(commits, commit)
	} else {
		for _, docID := range docIDs {
			commit, err := n.dagBlockToNodeDoc(dagBlock, docID)
			if err != nil {
				return false, err
			}
			commits = append(commits, commit)
		}
	}

	// if this is a time travel query or a _commits
	// (cid + undefined depth + docId) then we need to make sure the
	// target block actually belongs to the doc, since we are
	// bypassing the HeadFetcher for the first cid
	currentDocID := n.commitSelect.DocumentMapping.FirstOfName(commits[0], request.DocIDArgName)
	if n.commitSelect.Cids.HasValue() &&
		len(n.visitedNodes) == 0 &&
		n.commitSelect.DocIDs.HasValue() &&
		len(n.commitSelect.DocIDs.Value()) > 0 &&
		// todo - for now we just take the first docID and ignore the rest, an error
		// should be thrown in the parser anyway if the user provides more than one.
		// https://github.com/sourcenetwork/defradb/issues/4302
		currentDocID != n.commitSelect.DocIDs.Value()[0] {
		return false, ErrIncorrectOrMissingCID
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
	if (!n.commitSelect.Depth.HasValue() && !n.commitSelect.Cids.HasValue()) ||
		(n.commitSelect.Depth.HasValue() && n.depthVisited < n.commitSelect.Depth.Value()) {
		// Insert the newly fetched cids into the slice of queued items, in reverse order
		// so that the last new cid will be at the front of the slice
		n.queuedCids = append(make([]*cid.Cid, len(dagBlock.Heads)), n.queuedCids...)

		for i, head := range dagBlock.Heads {
			n.queuedCids[len(dagBlock.Heads)-i-1] = &head.Cid
		}
	}

	n.currentValue = commits[0]
	n.queuedCommits = commits[1:]
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

func (n *dagScanNode) dagBlockToNodeDoc(block *coreblock.Block, docID string) (core.Doc, error) {
	commit := n.commitSelect.DocumentMapping.NewDoc()
	link, err := block.GenerateLink()
	if err != nil {
		return core.Doc{}, err
	}
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.CidFieldName, link.String())

	collectionVersionId := block.Delta.GetCollectionVersionID()
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.CollectionVersionIDFieldName, collectionVersionId)

	// Pass the requester's identity so the collection lookup is authorised as them (rather than
	// anonymously) when node acp gates the get-collection operation.
	getColOpts := options.GetCollections().SetGetInactive(true).SetVersionID(collectionVersionId)
	if n.planner.identity.HasValue() {
		getColOpts = getColOpts.SetIdentity(n.planner.identity.Value())
	}
	cols, err := n.planner.db.GetCollections(n.planner.ctx, getColOpts)
	if err != nil {
		return core.Doc{}, err
	}
	if len(cols) == 0 {
		return core.Doc{}, client.NewErrCollectionNotFoundForCollectionVersion(collectionVersionId)
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

	if block.Signature != nil &&
		n.commitSelect.DocumentMapping.IndexesByName[request.SignatureFieldName] != nil {
		err := n.addSignatureFieldToDoc(*block.Signature, &commit)
		if err != nil {
			return core.Doc{}, err
		}
	}

	prio := block.Delta.GetPriority()
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.HeightFieldName, int64(prio))
	n.commitSelect.DocumentMapping.SetFirstOfName(&commit, request.FieldNameName, fieldName)

	if docID != "" {
		n.commitSelect.DocumentMapping.SetFirstOfName(
			&commit,
			request.DocIDArgName,
			docID,
		)
	}

	// scan for links
	linkedCids := make([]*cid.Cid, len(block.Links))
	for i, c := range block.Links {
		linkedCids[i] = &c.Cid
	}

	err = n.addLinksFieldToDoc(request.LinksFieldName, linkedCids, &commit)
	if err != nil {
		return core.Doc{}, err
	}

	// scan for heads
	blockCids := make([]*cid.Cid, len(block.Heads))
	for i, c := range block.Heads {
		blockCids[i] = &c.Cid
	}
	err = n.addLinksFieldToDoc(request.HeadsFieldName, blockCids, &commit)
	if err != nil {
		return core.Doc{}, err
	}

	return commit, nil
}

// addLinksFieldToDoc is responsible for adding both "links" and "heads" to a commit doc, since the
// only difference between the two is their link name.
func (n *dagScanNode) addLinksFieldToDoc(linksField string, links []*cid.Cid, commit *core.Doc) error {
	var dagScanNodes []*dagScanNode
	switch linksField {
	case request.LinksFieldName:
		dagScanNodes = n.linksScanNodes
	case request.HeadsFieldName:
		dagScanNodes = n.headsScanNodes
	}

	mappingIndexes := n.commitSelect.DocumentMapping.IndexesByName[linksField]
	for i, linksIndex := range mappingIndexes {
		// reset linkScanNode
		dagScanNodes[i].reset()
		dagScanNodes[i].queuedCids = links
		if parentDocID, ok := n.commitSelect.FirstOfName(*commit, request.DocIDArgName).(string); ok && parentDocID != "" {
			dagScanNodes[i].activeDocID = immutable.Some(parentDocID)
		}
		links := make([]core.Doc, 0)
		for {
			next, err := dagScanNodes[i].Next()
			if err != nil {
				return err
			}
			if !next {
				break
			}

			link := dagScanNodes[i].Value()
			pass, err := mapper.RunFilter(link, dagScanNodes[i].commitSelect.Filter)
			if err != nil {
				return err
			}

			if !pass {
				continue
			}

			links = append(links, link)
		}
		commit.Fields[linksIndex] = links
	}

	return nil
}

// addSignatureFieldToDoc adds the signature from the provided block link
// to the provided document.
// pre-condition: the signature needs to have been requested in the query selection
// so that it properly populates the mapper, will panic otherwise.
func (n *dagScanNode) addSignatureFieldToDoc(link cidlink.Link, commit *core.Doc) error {
	txn := datastore.CtxMustGetTxn(n.planner.ctx)

	sigIPLDBlock, err := txn.Blockstore().Get(n.planner.ctx, link.Cid)
	if err != nil {
		return NewErrGetSigBlock(err, link.Cid.String())
	}

	sigBlock, err := coreblock.GetSignatureBlockFromBytes(sigIPLDBlock.RawData())
	if err != nil {
		return NewErrDecodeSigBlock(err, link.Cid.String())
	}
	sigFieldIndexes := n.commitSelect.DocumentMapping.IndexesByName[request.SignatureFieldName]
	sigMapping := n.commitSelect.DocumentMapping.ChildMappings[sigFieldIndexes[0]]

	sigDoc := sigMapping.NewDoc()
	sigMapping.SetFirstOfName(&sigDoc, request.SignatureTypeFieldName, sigBlock.Header.Type)
	// Converting to a string from bytes[] results in it being presented as hex instead of base64
	sigMapping.SetFirstOfName(&sigDoc, request.SignatureIdentityFieldName, string(sigBlock.Header.Identity))
	sigMapping.SetFirstOfName(&sigDoc, request.SignatureValueFieldName, sigBlock.Value)

	n.commitSelect.DocumentMapping.SetFirstOfName(commit, request.SignatureFieldName, sigDoc)

	return nil
}

func (n *dagScanNode) reset() {
	n.visitedNodes = make(map[string]bool)
	n.queuedCids = make([]*cid.Cid, 0)
	n.queuedCommits = nil
	n.depthVisited = 0
	n.currentValue = core.Doc{}
	n.noResults = false
	n.activeDocID = immutable.None[string]()
	n.activeDocShortID = immutable.None[uint64]()
}

func (n *dagScanNode) setActiveDocIDFromCurrentHead() {
	n.activeDocID = immutable.None[string]()
	n.activeDocShortID = immutable.None[uint64]()

	currentKey := n.fetcher.CurrentKey()
	if !currentKey.HasValue() {
		return
	}

	docKey, ok := currentKey.Value().(keys.HeadstoreDocKey)
	if !ok {
		return
	}
	n.activeDocShortID = immutable.Some(docKey.DocShortID)
}

// commitDocIDs returns the documents a block's commit should be reported under.
//
// Most blocks resolve to a single document: the document established by the traversal (the
// docID inherited from a composite parent, or the active head's short ID), or a composite
// genesis block whose CID is, by construction, the public DocID. A field block can be shared
// by several documents (identical genesis field deltas produce one CID); for a CID-rooted
// query where no document is known, the block-CID owner index is consulted:
//   - a query that names a docID is narrowed to that document when it owns the block;
//   - a block with a single owner resolves to that owner;
//   - a `commits(cid:)` query (no docID) against a block with several owners fans the commit
//     out across every owner.
//
// The owner index must NOT be consulted before the traversal context: a shared field block
// would otherwise resolve to an arbitrary owner, mis-attributing the commit and dropping it
// from every other owner's history.
//
// An empty result means the block has no associated document (e.g. a collection-level delta)
// and yields a single commit with no docID.
func (n *dagScanNode) commitDocIDs(
	block *coreblock.Block,
	blockCID cid.Cid,
) ([]string, error) {
	if block.Delta.IsCollection() {
		return nil, nil
	}

	if block.Delta.IsComposite() && len(block.Heads) == 0 {
		docID := client.NewDocIDV0(blockCID).String()
		n.activeDocID = immutable.Some(docID)
		return []string{docID}, nil
	}

	if n.activeDocID.HasValue() {
		return []string{n.activeDocID.Value()}, nil
	}
	if n.activeDocShortID.HasValue() {
		docID, _, err := id.GetDocID(n.planner.ctx, n.activeDocShortID.Value())
		if err != nil {
			return nil, err
		}
		if docID != "" {
			n.activeDocID = immutable.Some(docID)
			return []string{docID}, nil
		}
	}

	owners, err := id.GetDocIDsForBlockFromStore(
		n.planner.ctx,
		datastore.CtxMustGetTxn(n.planner.ctx).Systemstore(),
		blockCID,
	)
	if err != nil {
		return nil, err
	}
	if len(owners) == 0 {
		return nil, nil
	}

	// A named docID narrows a shared block to that document.
	if n.commitSelect.DocIDs.HasValue() && len(n.commitSelect.DocIDs.Value()) > 0 {
		requested := n.commitSelect.DocIDs.Value()[0]
		for _, owner := range owners {
			if owner == requested {
				n.activeDocID = immutable.Some(owner)
				return []string{owner}, nil
			}
		}
		// The requested document does not own this block; let the caller treat it as a
		// cid/docID mismatch.
		return nil, nil
	}

	if len(owners) == 1 {
		n.activeDocID = immutable.Some(owners[0])
	}
	return owners, nil
}
