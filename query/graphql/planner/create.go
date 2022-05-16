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
	"encoding/json"
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/mapper"
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

	// newDoc is the JSON string of the new document, unparsed
	newDocStr string
	doc       *client.Document

	err error

	returned  bool
	selection *mapper.Select
}

func (n *createNode) Kind() string { return "createNode" }

func (n *createNode) Init() error { return nil }

func (n *createNode) Start() error {
	// parse the doc
	if n.newDocStr == "" {
		return fmt.Errorf("Invalid document to create")
	}

	doc, err := client.NewDocFromJSON([]byte(n.newDocStr))
	if err != nil {
		n.err = err
		return err
	}
	n.doc = doc
	return nil
}

// Next only returns once.
func (n *createNode) Next() (bool, error) {
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
		// On create the document will have no aliased fields/aggregates/etc so we can safely take
		// the first index.
		n.documentMapping.SetFirstOfName(&currentValue, i.Name(), value.Value())
	}

	n.returned = true
	n.currentValue = currentValue
	return true, nil
}

func (n *createNode) Spans(spans core.Spans) { /* no-op */ }

func (n *createNode) Close() error { return nil }

func (n *createNode) Source() planNode { return nil }

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *createNode) Explain() (map[string]interface{}, error) {
	data := map[string]interface{}{}
	err := json.Unmarshal([]byte(n.newDocStr), &data)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		dataLabel: data,
	}, nil
}

func (p *Planner) CreateDoc(parsed *mapper.Mutation) (planNode, error) {
	// create a mutation createNode.
	create := &createNode{
		p:         p,
		newDocStr: parsed.Data,
		selection: &parsed.Select,
		docMapper: docMapper{&parsed.DocumentMapping},
	}

	// get collection
	col, err := p.db.GetCollectionByName(p.ctx, parsed.Name)
	if err != nil {
		return nil, err
	}
	create.collection = col
	return p.SelectFromSource(&parsed.Select, create, true, nil)
}
