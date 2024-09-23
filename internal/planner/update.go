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
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

type updateNode struct {
	documentIterator
	docMapper

	p *Planner

	collection client.Collection

	filter *mapper.Filter

	docIDs []string

	// input map of fields and values
	input map[string]any

	isUpdating bool

	results planNode

	execInfo updateExecInfo
}

type updateExecInfo struct {
	// Total number of times updateNode was executed.
	iterations uint64

	// Total number of successful updates.
	updates uint64
}

// Next only returns once.
func (n *updateNode) Next() (bool, error) {
	n.execInfo.iterations++

	if n.isUpdating {
		for {
			next, err := n.results.Next()
			if err != nil {
				return false, err
			}
			if !next {
				break
			}

			n.currentValue = n.results.Value()

			docID, err := client.NewDocIDFromString(n.currentValue.GetID())
			if err != nil {
				return false, err
			}
			doc, err := n.collection.Get(n.p.ctx, docID, false)
			if err != nil {
				return false, err
			}
			for k, v := range n.input {
				if err := doc.Set(k, v); err != nil {
					return false, err
				}
			}
			err = n.collection.Update(n.p.ctx, doc)
			if err != nil {
				return false, err
			}

			n.execInfo.updates++
		}
		n.isUpdating = false

		// Re-init the results node, so that they can be properly yielded with the updated
		// values, as well as any formatting (e.g. aggregates, groupings, etc)
		err := n.results.Init()
		if err != nil {
			return false, err
		}
	}

	next, err := n.results.Next()
	if err != nil {
		return false, err
	}
	if !next {
		return false, nil
	}

	n.currentValue = n.results.Value()
	return true, nil
}

func (n *updateNode) Kind() string { return "updateNode" }

func (n *updateNode) Spans(spans core.Spans) { n.results.Spans(spans) }

func (n *updateNode) Init() error { return n.results.Init() }

func (n *updateNode) Start() error {
	return n.results.Start()
}

func (n *updateNode) Close() error {
	return n.results.Close()
}

func (n *updateNode) Source() planNode { return n.results }

func (n *updateNode) simpleExplain() (map[string]any, error) {
	simpleExplainMap := map[string]any{}

	// Add the document id(s) that request wants to update.
	simpleExplainMap[request.DocIDArgName] = n.docIDs

	// Add the filter attribute if it exists, otherwise have it nil.
	if n.filter == nil {
		simpleExplainMap[filterLabel] = nil
	} else {
		simpleExplainMap[filterLabel] = n.filter.ToMap(n.documentMapping)
	}

	// Add the attribute that represents the patch to update with.
	simpleExplainMap[inputLabel] = n.input

	return simpleExplainMap, nil
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *updateNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return n.simpleExplain()

	case request.ExecuteExplain:
		return map[string]any{
			"iterations": n.execInfo.iterations,
			"updates":    n.execInfo.updates,
		}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}

func (p *Planner) UpdateDocs(parsed *mapper.Mutation) (planNode, error) {
	update := &updateNode{
		p:          p,
		filter:     parsed.Filter,
		docIDs:     parsed.DocIDs.Value(),
		isUpdating: true,
		docMapper:  docMapper{parsed.DocumentMapping},
	}

	// update mutation only supports a single input
	if len(parsed.Input) > 0 {
		update.input = parsed.Input[0]
	}

	// get collection
	col, err := p.db.GetCollectionByName(p.ctx, parsed.Name)
	if err != nil {
		return nil, err
	}
	update.collection = col

	// create the results Select node
	resultsNode, err := p.Select(&parsed.Select)
	if err != nil {
		return nil, err
	}
	update.results = resultsNode

	return update, nil
}
