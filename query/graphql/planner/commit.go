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
	"math"

	cid "github.com/ipfs/go-cid"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

type commitSelectNode struct {
	p *Planner

	source *dagScanNode

	subRenderInfo map[string]renderInfo

	doc map[string]interface{}
}

func (n *commitSelectNode) Init() error {
	return n.source.Init()
}

func (n *commitSelectNode) Start() error {
	return n.source.Start()
}

func (n *commitSelectNode) Next() (bool, error) {
	if next, err := n.source.Next(); !next {
		return false, err
	}

	n.doc = n.source.Values()
	return true, nil
}

func (n *commitSelectNode) Values() map[string]interface{} {
	return n.doc
}

func (n *commitSelectNode) Spans(spans core.Spans) {
	n.source.Spans(spans)
}

func (n *commitSelectNode) Close() error {
	return n.source.Close()
}

func (n *commitSelectNode) Source() planNode {
	return n.source
}

// AppendNode implements appendNode
// func (n *commitSelectNode) AppendNode() bool {
// 	return true
// }

func (p *Planner) CommitSelect(parsed *parser.CommitSelect) (planNode, error) {
	// check type of commit select (all, latest, one)
	var commit *commitSelectNode
	var err error
	switch parsed.Type {
	case parser.LatestCommits:
		commit, err = p.commitSelectLatest(parsed)
	case parser.OneCommit:
		commit, err = p.commitSelectBlock(parsed)
	case parser.AllCommits:
		commit, err = p.commitSelectAll(parsed)
	default:
		return nil, fmt.Errorf("Invalid CommitSelect type")
	}
	if err != nil {
		return nil, err
	}
	slct := parsed.ToSelect()
	plan, err := p.SelectFromSource(slct, commit, false, nil)
	if err != nil {
		return nil, err
	}
	return &commitSelectTopNode{
		p:    p,
		plan: plan,
	}, nil
}

// commitSelectLatest is a CommitSelect node initalized with a headsetScanNode and a DocKey
func (p *Planner) commitSelectLatest(parsed *parser.CommitSelect) (*commitSelectNode, error) {
	dag := p.DAGScan()
	headset := p.HeadScan()
	// @todo: Get Collection field ID
	if parsed.FieldName == "" {
		parsed.FieldName = core.COMPOSITE_NAMESPACE
	}
	dag.field = parsed.FieldName
	if parsed.DocKey != "" {
		key := core.DataStoreKey{}.WithDocKey(parsed.DocKey).WithFieldId(parsed.FieldName)
		headset.key = key
	}
	dag.headset = headset
	commit := &commitSelectNode{
		p:             p,
		source:        dag,
		subRenderInfo: make(map[string]renderInfo),
	}

	return commit, nil
}

// commitSelectBlock is a CommitSelect node intialized witout a headsetScanNode, and is
// expected to be given a target CID in the parser.CommitSelect object. It returns
// a single commit if found
func (p *Planner) commitSelectBlock(parsed *parser.CommitSelect) (*commitSelectNode, error) {
	dag := p.DAGScan()
	if parsed.Cid != "" {
		c, err := cid.Decode(parsed.Cid)
		if err != nil {
			return nil, err
		}
		dag.cid = &c
	} // @todo: handle error if no CID is given

	return &commitSelectNode{
		p:             p,
		source:        dag,
		subRenderInfo: make(map[string]renderInfo),
	}, nil
}

// commitSelectAll is a CommitSelect initialized with a headsetScanNode, and will
// recursively return all graph commits in order.
func (p *Planner) commitSelectAll(parsed *parser.CommitSelect) (*commitSelectNode, error) {
	dag := p.DAGScan()
	headset := p.HeadScan()
	// @todo: Get Collection field ID
	if parsed.FieldName == "" {
		parsed.FieldName = core.COMPOSITE_NAMESPACE
	}
	dag.field = parsed.FieldName
	if parsed.DocKey != "" {
		key := core.DataStoreKey{}.WithDocKey(parsed.DocKey).WithFieldId(parsed.FieldName)
		headset.key = key
	}
	dag.headset = headset
	dag.depthLimit = math.MaxUint32 // infinite depth
	// dag.key = &key
	commit := &commitSelectNode{
		p:             p,
		source:        dag,
		subRenderInfo: make(map[string]renderInfo),
	}

	return commit, nil
}

// commitSelectTopNode is a wrapper for the selectTopNode
// in the case where the select is actually a CommitSelect
type commitSelectTopNode struct {
	p    *Planner
	plan planNode
}

func (n *commitSelectTopNode) Init() error                    { return n.plan.Init() }
func (n *commitSelectTopNode) Start() error                   { return n.plan.Start() }
func (n *commitSelectTopNode) Next() (bool, error)            { return n.plan.Next() }
func (n *commitSelectTopNode) Spans(spans core.Spans)         { n.plan.Spans(spans) }
func (n *commitSelectTopNode) Values() map[string]interface{} { return n.plan.Values() }
func (n *commitSelectTopNode) Source() planNode               { return n.plan }
func (n *commitSelectTopNode) Close() error {
	if n.plan == nil {
		return nil
	}
	return n.plan.Close()
}

func (n *commitSelectTopNode) Append() bool { return true }
