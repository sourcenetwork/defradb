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
			key, err2 := client.NewDocKeyFromString(n.ids[0])
			if err2 != nil {
				return false, err2
			}
			results, err = n.collection.DeleteWithKey(n.p.ctx, key)
		} else if numids > 1 {
			keys := make([]client.DocKey, len(n.ids))
			for i, v := range n.ids {
				keys[i], err = client.NewDocKeyFromString(v)
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
			err := n.deleteIter.docs.AddDoc(core.Doc{"_key": resKey})
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

func (n *deleteNode) Value() core.Doc {
	return n.deleteIter.Value()
}

func (n *deleteNode) Spans(spans core.Spans) {
	/* no-op */
}

func (n *deleteNode) Kind() string {
	return "deleteNode"
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

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *deleteNode) Explain() (map[string]interface{}, error) {
	explainerMap := map[string]interface{}{}

	// Add the document id(s) that request wants to delete.
	explainerMap[idsLabel] = n.ids

	// Add the filter attribute if it exists, otherwise have it nil.
	if n.filter == nil || n.filter.Conditions == nil {
		explainerMap[filterLabel] = nil
	} else {
		explainerMap[filterLabel] = n.filter.Conditions
	}

	return explainerMap, nil
}

func (p *Planner) DeleteDocs(parsed *parser.Mutation) (planNode, error) {
	delete := &deleteNode{
		p:          p,
		filter:     parsed.Filter,
		ids:        parsed.IDs,
		isDeleting: true,
	}

	// get collection
	col, err := p.db.GetCollectionByName(p.ctx, parsed.Schema)
	if err != nil {
		return nil, err
	}

	delete.collection = col.WithTxn(p.txn)

	slct := parsed.ToSelect()
	return p.SelectFromSource(slct, delete, true, nil)
}
