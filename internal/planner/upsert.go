// Copyright 2024 Democratized Data Foundation
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

type upsertNode struct {
	documentIterator
	docMapper

	p             *Planner
	collection    client.Collection
	filter        *mapper.Filter
	createInput   map[string]any
	updateInput   map[string]any
	isInitialized bool
	source        planNode
}

// Next only returns once.
func (n *upsertNode) Next() (bool, error) {
	if !n.isInitialized {
		next, err := n.source.Next()
		if err != nil {
			return false, err
		}
		if next {
			n.currentValue = n.source.Value()
			// make sure multiple documents do not match
			next, err := n.source.Next()
			if err != nil {
				return false, err
			}
			if next {
				return false, ErrUpsertMultipleDocuments
			}
			docID, err := client.NewDocIDFromString(n.currentValue.GetID())
			if err != nil {
				return false, err
			}
			doc, err := n.collection.Get(n.p.ctx, docID, false)
			if err != nil {
				return false, err
			}
			for k, v := range n.updateInput {
				if err := doc.Set(k, v); err != nil {
					return false, err
				}
			}
			err = n.collection.Update(n.p.ctx, doc)
			if err != nil {
				return false, err
			}
		} else {
			doc, err := client.NewDocFromMap(n.createInput, n.collection.Definition())
			if err != nil {
				return false, err
			}
			err = n.collection.Create(n.p.ctx, doc)
			if err != nil {
				return false, err
			}
			n.source.Spans(docIDsToSpans(documentsToDocIDs(doc), n.collection.Description()))
		}
		err = n.source.Init()
		if err != nil {
			return false, err
		}
		n.isInitialized = true
	}
	next, err := n.source.Next()
	if err != nil {
		return false, err
	}
	if !next {
		return false, nil
	}
	n.currentValue = n.source.Value()
	return true, nil
}

func (n *upsertNode) Kind() string {
	return "upsertNode"
}

func (n *upsertNode) Spans(spans []core.Span) {
	n.source.Spans(spans)
}

func (n *upsertNode) Init() error {
	return n.source.Init()
}

func (n *upsertNode) Start() error {
	return n.source.Start()
}

func (n *upsertNode) Close() error {
	return n.source.Close()
}

func (n *upsertNode) Source() planNode {
	return n.source
}

func (n *upsertNode) simpleExplain() (map[string]any, error) {
	simpleExplainMap := map[string]any{}

	// Add the filter attribute
	simpleExplainMap[filterLabel] = n.filter.ToMap(n.documentMapping)

	// Add the attribute that represents the values to create or update.
	simpleExplainMap[updateInputLabel] = n.updateInput
	simpleExplainMap[createInputLabel] = n.createInput

	return simpleExplainMap, nil
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *upsertNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return n.simpleExplain()

	case request.ExecuteExplain:
		return map[string]any{}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}

func (p *Planner) UpsertDocs(parsed *mapper.Mutation) (planNode, error) {
	upsert := &upsertNode{
		p:           p,
		filter:      parsed.Filter,
		updateInput: parsed.UpdateInput,
		docMapper:   docMapper{parsed.DocumentMapping},
	}

	if len(parsed.CreateInput) > 0 {
		upsert.createInput = parsed.CreateInput[0]
	}

	// get collection
	col, err := p.db.GetCollectionByName(p.ctx, parsed.Name)
	if err != nil {
		return nil, err
	}
	upsert.collection = col

	// create the results Select node
	resultsNode, err := p.Select(&parsed.Select)
	if err != nil {
		return nil, err
	}
	upsert.source = resultsNode

	return upsert, nil
}
