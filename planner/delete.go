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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

type deleteNode struct {
	documentIterator
	docMapper

	p *Planner

	collection client.Collection
	source     planNode

	filter *mapper.Filter
	docIDs []string

	execInfo deleteExecInfo
}

type deleteExecInfo struct {
	// Total number of times deleteNode was executed.
	iterations uint64
}

func (n *deleteNode) Next() (bool, error) {
	n.execInfo.iterations++

	next, err := n.source.Next()
	if err != nil {
		return false, err
	}
	if !next {
		return false, nil
	}

	n.currentValue = n.source.Value()
	key, err := client.NewDocKeyFromString(n.currentValue.GetKey())
	if err != nil {
		return false, err
	}
	_, err = n.collection.DeleteWithKey(n.p.ctx, key)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (n *deleteNode) Spans(spans core.Spans) {
	n.source.Spans(spans)
}

func (n *deleteNode) Kind() string {
	return "deleteNode"
}

func (n *deleteNode) Init() error {
	return n.source.Init()
}

func (n *deleteNode) Start() error {
	return n.source.Start()
}

func (n *deleteNode) Close() error {
	return n.source.Close()
}

func (n *deleteNode) Source() planNode {
	return n.source
}

func (n *deleteNode) simpleExplain() (map[string]any, error) {
	simpleExplainMap := map[string]any{}

	// Add the document id(s) that request wants to delete.
	simpleExplainMap[request.DocIDs] = n.docIDs

	// Add the filter attribute if it exists, otherwise have it nil.
	if n.filter == nil {
		simpleExplainMap[filterLabel] = nil
	} else {
		simpleExplainMap[filterLabel] = n.filter.ToMap(n.documentMapping)
	}

	return simpleExplainMap, nil
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *deleteNode) Explain(explainType request.ExplainType) (map[string]any, error) {
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

func (p *Planner) DeleteDocs(parsed *mapper.Mutation) (planNode, error) {
	col, err := p.db.GetCollectionByName(p.ctx, parsed.Name)
	if err != nil {
		return nil, err
	}

	slctNode, err := p.Select(&parsed.Select)
	if err != nil {
		return nil, err
	}

	return &deleteNode{
		p:          p,
		filter:     parsed.Filter,
		docIDs:     parsed.DocKeys.Value(),
		collection: col.WithTxn(p.txn),
		source:     slctNode,
		docMapper:  docMapper{parsed.DocumentMapping},
	}, nil
}
