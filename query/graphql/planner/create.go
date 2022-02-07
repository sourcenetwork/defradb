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
	"errors"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/document"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

// createNode is used to construct and execute
// an object create mutation.
//
// Create nodes are the simplest of the object mutations
// Each Iteration of the plan, creates and returns one
// document, until we've exhausted the payload. No filtering
// or Select plans
type createNode struct {
	p *Planner

	// cache information about the original data source
	// collection name, meta-data, etc.
	collection client.Collection

	// newDoc is the JSON string of the new document, unparsed
	newDocStr string
	doc       *document.Document

	err error

	returned bool
}

func (n *createNode) Init() error { return nil }

func (n *createNode) Start() error {
	// parse the doc
	if n.newDocStr == "" {
		return errors.New("Invalid document to create")
	}

	doc, err := document.NewFromJSON([]byte(n.newDocStr))
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

	n.returned = true
	return true, nil
}

func (n *createNode) Spans(spans core.Spans) { /* no-op */ }

func (n *createNode) Values() map[string]interface{} {
	val, _ := n.doc.ToMap()
	return val
}

func (n *createNode) Close() error { return nil }

func (n *createNode) Source() planNode { return nil }

func (p *Planner) CreateDoc(parsed *parser.Mutation) (planNode, error) {
	// create a mutation createNode.
	create := &createNode{
		p:         p,
		newDocStr: parsed.Data,
	}

	// get collection
	col, err := p.db.GetCollection(p.ctx, parsed.Schema)
	if err != nil {
		return nil, err
	}
	create.collection = col

	// last step, create a basic Select statement
	// from the parsed Mutation object
	// and construct a new Select planNode
	// which uses the new create node as its
	// source, instead of a scan node.
	slct := parsed.ToSelect()
	return p.SelectFromSource(slct, create, true, nil)
}
