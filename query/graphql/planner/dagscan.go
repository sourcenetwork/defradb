package planner

import (
	// "errors"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/fetcher"

	"github.com/fxamacker/cbor/v2"
	blocks "github.com/ipfs/go-block-format"
	cid "github.com/ipfs/go-cid"
	dag "github.com/ipfs/go-merkledag"
	"github.com/pkg/errors"
)

type headsetScanNode struct {
	p   *Planner
	key core.Key

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

	key *core.Key
	cid *cid.Cid

	depthLimit int
	headset    *headsetScanNode

	// previousScanNode planNode
	// linksScanNode    planNode

	block blocks.Block
}

func (p *Planner) DAGScan() *dagScanNode {
	return &dagScanNode{p: p}
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
		n.headset.Spans(spans)
	} else {
		data := spans[0].Start().String()
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
	if n.cid == nil {
		if n.headset == nil {
			return false, errors.New("DAG Scan node has no target, missing key, cid, or headset scan node")
		}
		if next, err := n.headset.Next(); !next {
			return false, err
		}

		val := n.headset.Values()
		cid, ok := val["cid"].(cid.Cid)
		if !ok {
			return false, errors.New("Headset scan node returned an invalid cid")
		}
		n.cid = &cid
	}

	// use the stored cid to scan through the blockstore
	// clear the cid after
	store := n.p.txn.DAGstore()
	block, err := store.Get(*n.cid)
	if err != nil { // handle error?
		return false, err
	}
	n.block = block
	n.cid = nil // clear cid for next round
	return true, nil
}

// func (n *dagScanNode) nextHead() (cid.Cid, error) {

// }

func (n *dagScanNode) Values() map[string]interface{} {
	doc, _ := dagNodeToCommitMap(n.block)
	// if err != nil { // handle error?
	// 	return map[string]interface{}{
	// 		"error": err.Error(),
	// 	}
	// }
	return doc
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

func dagNodeToCommitMap(block blocks.Block) (map[string]interface{}, error) {
	commit := map[string]interface{}{
		"cid": block.Cid().String(),
	}

	// decode the delta, get the priority and payload
	nd, err := dag.DecodeProtobuf(block.RawData())
	if err != nil {
		return nil, err
	}

	var delta map[string]interface{}
	if err := cbor.Unmarshal(nd.Data(), &delta); err != nil {
		return nil, err
	}

	prio, ok := delta["Priority"].(uint64)
	if !ok {
		return nil, errors.New("Commit Delta missing priority key")
	}

	commit["height"] = int64(prio)
	commit["delta"] = delta["Data"] // check

	// links
	links := make([]map[string]interface{}, len(nd.Links()))
	for i, l := range nd.Links() {
		link := map[string]interface{}{
			"name": l.Name,
			"cid":  l.Cid.String(),
		}
		links[i] = link
	}
	commit["links"] = links
	return commit, nil
}
