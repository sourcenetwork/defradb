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
	"encoding/json"
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/document/key"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

type deleteNode struct {
	p *Planner

	collection client.Collection

	filter *parser.Filter
	ids    []string

	patch string

	isUpdating bool
	deleteIter *valuesNode

	results planNode
}

// Next only returns once.
func (n *deleteNode) Next() (bool, error) {
	// if err := n.collection.WithTxn(n.p.txn).Create(n.doc); err != nil {
	// 	return false, err
	// }

	if n.isUpdating {
		// create our result values node
		if n.deleteIter == nil {
			vnode := n.p.newContainerValuesNode(nil)
			n.deleteIter = vnode
		}

		// apply the deletes
		// @todo: handle filter vs ID based
		var results *client.DeleteResult
		var err error
		numids := len(n.ids)
		if numids == 1 {
			fmt.Println("single key")
			key, err2 := key.NewFromString(n.ids[0])
			if err2 != nil {
				return false, err2
			}
			results, err = n.collection.DeleteWithKey(n.p.ctx, key, n.patch)
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
			results, err = n.collection.DeleteWithKeys(n.p.ctx, keys, n.patch)
		} else {
			fmt.Println("filter")
			results, err = n.collection.DeleteWithFilter(n.p.ctx, n.filter, n.patch)
		}

		fmt.Println("delete node error:", err)
		if err != nil {
			return false, err
		}

		// consume the deletes into our valuesNode
		fmt.Println(results)
		for _, resKey := range results.DocKeys {
			n.deleteIter.docs.AddDoc(map[string]interface{}{"_key": resKey})
		}
		n.isUpdating = false

		// lets release the results dockeys slice memory
		results.DocKeys = nil
	}

	// next, err := n.deleteIter.Next()
	// if !next {
	// 	return false, err
	// }
	return n.deleteIter.Next()
}

func (n *deleteNode) Values() map[string]interface{} {
	deletedDoc := n.deleteIter.Values()
	// create a new span with the deleteDoc._key
	docKeyStr := deletedDoc["_key"].(string)
	desc := n.collection.Description()
	deletedDocKeyIndex := base.MakeIndexKey(&desc, &desc.Indexes[0], core.NewKey(docKeyStr))
	spans := core.Spans{core.NewSpan(deletedDocKeyIndex, deletedDocKeyIndex.PrefixEnd())}

	n.results.Spans(spans)
	n.results.Init()

	// get the next result based on our point lookup
	next, err := n.results.Next()
	if !next || err != nil {
		panic(err) //handle better?
	}

	// we're only expecting a single value from our pointlookup
	return n.results.Values()
}

func (n *deleteNode) Spans(spans core.Spans) { /* no-op */ }
func (n *deleteNode) Init() error            { return nil }

func (n *deleteNode) Start() error {

	//println("\n-------------------------------------------------------")
	//test := n.results.Values
	//ssTest, _ := json.MarshalIndent(test, "", "\t")
	//fmt.Println(string(ssTest))
	//println("=======================================================")
	//fmt.Printf("%#v", test)
	//println("\n-------------------------------------------------------")

	return n.results.Start()
}

func (n *deleteNode) Close() error {
	return n.results.Close()
}

func (n *deleteNode) Source() planNode { return nil }

func (p *Planner) DeleteDocs(parsed *parser.Mutation) (planNode, error) {
	delete := &deleteNode{
		p:          p,
		filter:     parsed.Filter,
		ids:        parsed.IDs,
		isUpdating: true,
		patch:      parsed.Data,
	}

	// println("\n DELETE - delete: -------------------------------------------------------")
	// test := delete
	// ssTest, _ := json.MarshalIndent(test, "", "\t")
	// fmt.Println(string(ssTest))
	// println("=======================================================")
	// fmt.Printf("%#v", test)
	// println("\n-------------------------------------------------------")

	// get collection
	col, err := p.db.GetCollection(p.ctx, parsed.Schema)

	// println("\n DELETE - col: -------------------------------------------------------")
	// test1 := col
	// ssTest1, _ := json.MarshalIndent(test1, "", "\t")
	// fmt.Println(string(ssTest1))
	// println("=======================================================")
	// fmt.Printf("%#v", test1)
	// println("\n-------------------------------------------------------")

	if err != nil {
		return nil, err
	}
	delete.collection = col.WithTxn(p.txn)

	// println("\n DELETE - collection: -------------------------------------------------------")
	// test1 := delete.collection
	// ssTest1, _ := json.MarshalIndent(test1, "", "\t")
	// fmt.Println(string(ssTest1))
	// println("=======================================================")
	// fmt.Printf("%#v", test1)
	// println("\n-------------------------------------------------------")

	// create the results Select node
	slct := parsed.ToSelect()
	slctNode, err := p.Select(slct)

	println("\n DELETE - select: -------------------------------------------------------")
	test1 := slctNode
	ssTest1, _ := json.MarshalIndent(test1, "", "\t")
	fmt.Println(string(ssTest1))
	println("=======================================================")
	fmt.Printf("%#v", test1)
	println("\n-------------------------------------------------------")

	if err != nil {
		println("\nXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX")
		return nil, err
	}
	delete.results = slctNode
	return delete, nil
}
