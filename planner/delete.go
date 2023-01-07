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
	ids    []string
}

func (n *deleteNode) Next() (bool, error) {
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

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *deleteNode) Explain() (map[string]any, error) {
	explainerMap := map[string]any{}

	// Add the document id(s) that request wants to delete.
	explainerMap[idsLabel] = n.ids

	// Add the filter attribute if it exists, otherwise have it nil.
	if n.filter == nil || n.filter.ExternalConditions == nil {
		explainerMap[filterLabel] = nil
	} else {
		explainerMap[filterLabel] = n.filter.ExternalConditions
	}

	return explainerMap, nil
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
		ids:        parsed.DocKeys.Value(),
		collection: col.WithTxn(p.txn),
		source:     slctNode,
		docMapper:  docMapper{&parsed.DocumentMapping},
	}, nil
}
