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
	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

type deleteNode struct {
	p *Planner

	collection client.Collection

	filter *parser.Filter
	ids    []string

	isDeleting bool
	deleteIter *valuesNode
}

// Next only returns once.
func (n *deleteNode) Next() (bool, error) {
	if n.isDeleting {
		// create our result values node
		if n.deleteIter == nil {
			vnode := n.p.newContainerValuesNode(nil)
			n.deleteIter = vnode
		}

		// Apply the deletes
		var results *client.DeleteResult
		var err error
		numids := len(n.ids)

		if n.filter != nil && numids != 0 {
			return false, errors.New("Error: can't use filter and id / ids together.")
		} else if n.filter != nil {
			results, err = n.collection.DeleteWithFilter(n.p.ctx, n.filter)
		} else if numids == 0 {
			return false, errors.New("Error: no id(s) provided while delete mutation.")
		} else if numids == 1 {
			key, err2 := key.NewFromString(n.ids[0])
			if err2 != nil {
				return false, err2
			}
			results, err = n.collection.DeleteWithKey(n.p.ctx, key)
		} else if numids > 1 {
			keys := make([]key.DocKey, len(n.ids))
			for i, v := range n.ids {
				keys[i], err = key.NewFromString(v)
				if err != nil {
					return false, err
				}
			}
			results, err = n.collection.DeleteWithKeys(n.p.ctx, keys)
		} else {
			return false, errors.New("Error: out of scope use of delete mutation.")
		}

		if err != nil {
			return false, err
		}

		// Consume the deletes into our valuesNode
		for _, resKey := range results.DocKeys {
			err := n.deleteIter.docs.AddDoc(map[string]interface{}{"_key": resKey})
			if err != nil {
				return false, err
			}
		}

		n.isDeleting = false

		// lets release the results dockeys slice memory
		results.DocKeys = nil
	}

	return n.deleteIter.Next()
}

func (n *deleteNode) Values() map[string]interface{} {
	return n.deleteIter.Values()
}

func (n *deleteNode) Spans(spans core.Spans) {
	/* no-op */
}

func (n *deleteNode) Init() error {
	return nil
}

func (n *deleteNode) Start() error {
	return nil
}

func (n *deleteNode) Close() error {
	return nil
}

func (n *deleteNode) Source() planNode {
	return nil
}

func (p *Planner) DeleteDocs(parsed *parser.Mutation) (planNode, error) {
	delete := &deleteNode{
		p:          p,
		filter:     parsed.Filter,
		ids:        parsed.IDs,
		isDeleting: true,
	}

	// get collection
	col, err := p.db.GetCollection(p.ctx, parsed.Schema)
	if err != nil {
		return nil, err
	}

	delete.collection = col.WithTxn(p.txn)

	slct := parsed.ToSelect()
	return p.SelectFromSource(slct, delete, true, nil)
}
