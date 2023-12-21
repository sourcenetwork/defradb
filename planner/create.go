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
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

// createNode is used to construct and execute
// an object create mutation.
//
// Create nodes are the simplest of the object mutations
// Each Iteration of the plan, creates and returns one
// document, until we've exhausted the payload. No filtering
// or Select plans
type createNode struct {
	documentIterator
	docMapper

	p *Planner

	// cache information about the original data source
	// collection name, meta-data, etc.
	collection client.Collection

	// input map of fields and values
	input map[string]any
	doc   *client.Document

	err error

	returned bool
	results  planNode

	execInfo createExecInfo
}

type createExecInfo struct {
	// Total number of times createNode was executed.
	iterations uint64
}

func (n *createNode) Kind() string { return "createNode" }

func (n *createNode) Init() error { return nil }

func (n *createNode) Start() error {
	doc, err := client.NewDocFromMap(n.input)
	if err != nil {
		n.err = err
		return err
	}
	n.doc = doc
	return nil
}

// Next only returns once.
func (n *createNode) Next() (bool, error) {
	n.execInfo.iterations++

	if n.err != nil {
		return false, n.err
	}

	if n.returned {
		return false, nil
	}

	if err := n.collection.WithTxn(n.p.txn).Create(n.p.ctx, n.doc); err != nil {
		return false, err
	}

	currentValue := n.documentMapping.NewDoc()

	currentValue.SetKey(n.doc.Key().String())
	for i, value := range n.doc.Values() {
		if len(n.documentMapping.IndexesByName[i.Name()]) > 0 {
			n.documentMapping.SetFirstOfName(&currentValue, i.Name(), value.Value())
		} else if aliasName := i.Name() + request.RelatedObjectID; len(n.documentMapping.IndexesByName[aliasName]) > 0 {
			n.documentMapping.SetFirstOfName(&currentValue, aliasName, value.Value())
		} else {
			return false, client.NewErrFieldNotExist(i.Name())
		}
	}

	n.returned = true
	n.currentValue = currentValue

	desc := n.collection.Description()
	docKey := base.MakeDocKey(desc, currentValue.GetKey())
	n.results.Spans(core.NewSpans(core.NewSpan(docKey, docKey.PrefixEnd())))

	err := n.results.Init()
	if err != nil {
		return false, err
	}

	err = n.results.Start()
	if err != nil {
		return false, err
	}

	// get the next result based on our point lookup
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

func (n *createNode) Spans(spans core.Spans) { /* no-op */ }

func (n *createNode) Close() error {
	return n.results.Close()
}

func (n *createNode) Source() planNode { return n.results }

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *createNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return map[string]any{
			dataLabel: n.input,
		}, nil

	case request.ExecuteExplain:
		return map[string]any{
			"iterations": n.execInfo.iterations,
		}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}

func (p *Planner) CreateDoc(parsed *mapper.Mutation) (planNode, error) {
	results, err := p.Select(&parsed.Select)
	if err != nil {
		return nil, err
	}

	// create a mutation createNode.
	create := &createNode{
		p:         p,
		input:     parsed.Input,
		results:   results,
		docMapper: docMapper{parsed.DocumentMapping},
	}

	// get collection
	col, err := p.db.GetCollectionByName(p.ctx, parsed.Name)
	if err != nil {
		return nil, err
	}
	create.collection = col
	return create, nil
}
