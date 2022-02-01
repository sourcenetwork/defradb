// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package planner

import (
	"fmt"

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

	results planNode
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
		if numids == 1 {
			fmt.Println("single key")
			key, err2 := key.NewFromString(n.ids[0])
			if err2 != nil {
				return false, err2
			}
			results, err = n.collection.DeleteWithKey(n.p.ctx, key)
		} else if numids > 1 {
			fmt.Println("multi key")
			// todo
			keys := make([]key.DocKey, len(n.ids))
			for i, v := range n.ids {
				keys[i], err = key.NewFromString(v)
				if err != nil {
					return false, err
				}
			}
			results, err = n.collection.DeleteWithKeys(n.p.ctx, keys)
		} else { // @todo: handle filter vs ID based
			fmt.Println("filter")
			results, err = n.collection.DeleteWithFilter(n.p.ctx, n.filter)
		}

		fmt.Println("delete node error:", err)
		if err != nil {
			return false, err
		}

		// Consume the deletes into our valuesNode
		fmt.Println(results)
		for _, resKey := range results.DocKeys {
			n.deleteIter.docs.AddDoc(map[string]interface{}{"_key": resKey})
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
	return n.results.Start()
}

func (n *deleteNode) Close() error {
	return n.results.Close()
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

	// create the results Select node
	slct := parsed.ToSelect()
	slctNode, err := p.Select(slct)
	if err != nil {
		return nil, err
	}
	delete.results = slctNode
	return delete, nil
}
