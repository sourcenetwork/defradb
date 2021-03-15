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
	"errors"
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
	n.renderDoc()
	return true, nil
}

func (n *commitSelectNode) Values() map[string]interface{} {
	return n.doc
}

func (n *commitSelectNode) Spans(spans core.Spans) {
	n.source.Spans(spans)
}

func (n *commitSelectNode) Close() {
	n.source.Close()
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
		return nil, errors.New("Invalid CommitSelect type")
	}
	if err != nil {
		return nil, err
	}
	err = commit.initFields(parsed)
	if err != nil {
		return nil, err
	}
	slct := parsed.ToSelect()
	plan, err := p.SelectFromSource(slct, commit, false)
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
		parsed.FieldName = "C" // C for composite DAG
	}
	dag.field = parsed.FieldName
	if parsed.DocKey != "" {
		key := core.NewKey(parsed.DocKey + "/" + parsed.FieldName)
		headset.key = key
	}
	dag.headset = headset
	// dag.depthLimit = 1
	// dag.key = &key
	commit := &commitSelectNode{
		p:             p,
		source:        dag,
		subRenderInfo: make(map[string]renderInfo),
	}

	return commit, nil
}

func (p *Planner) commitSelectBlock(parsed *parser.CommitSelect) (*commitSelectNode, error) {
	dag := p.DAGScan()
	if parsed.Cid != "" {
		c, err := cid.Decode(parsed.Cid)
		if err != nil {
			return nil, err
		}
		dag.cid = &c
		fmt.Println("got cid:", c)
	}

	return &commitSelectNode{
		p:             p,
		source:        dag,
		subRenderInfo: make(map[string]renderInfo),
	}, nil
}

func (p *Planner) commitSelectAll(parsed *parser.CommitSelect) (*commitSelectNode, error) {
	dag := p.DAGScan()
	headset := p.HeadScan()
	// @todo: Get Collection field ID
	if parsed.FieldName == "" {
		parsed.FieldName = "C" // C for composite DAG
	}
	dag.field = parsed.FieldName
	if parsed.DocKey != "" {
		key := core.NewKey(parsed.DocKey + "/" + parsed.FieldName)
		headset.key = key
	}
	dag.headset = headset
	dag.depthLimit = math.MaxUint32 // inifinite depth
	// dag.key = &key
	commit := &commitSelectNode{
		p:             p,
		source:        dag,
		subRenderInfo: make(map[string]renderInfo),
	}

	return commit, nil
}

// renderDoc applies the render meta-data to the
// links/previous sub selections for a commit type
// query.
func (n *commitSelectNode) renderDoc() {
	for subfield, info := range n.subRenderInfo {
		renderData := map[string]interface{}{
			"numResults": info.numResults,
			"fields":     info.fields,
			"aliases":    info.aliases,
		}
		for _, subcommit := range n.doc[subfield].([]map[string]interface{}) {
			subcommit["__render"] = renderData
		}

	}
}

func (n *commitSelectNode) initFields(parsed parser.Selection) error {
	for _, selection := range parsed.GetSelections() {
		switch node := selection.(type) {
		case *parser.Select:
			info := renderInfo{}
			for _, f := range node.Fields {
				info.fields = append(info.fields, f.GetName())
				info.aliases = append(info.aliases, f.GetAlias())
				info.numResults++
			}
			n.subRenderInfo[node.Name] = info
		}
	}
	return nil
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
func (n *commitSelectTopNode) Close() {
	if n.plan != nil {
		n.plan.Close()
	}
}

func (n *commitSelectTopNode) Append() bool { return true }
