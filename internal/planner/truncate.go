// Copyright 2026 Democratized Data Foundation
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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

type truncateNode struct {
	documentIterator
	docMapper

	p            *Planner
	collection   client.Collection
	filter       immutable.Option[map[string]any]
	pruneHistory bool
	isDone       bool
}

func (n *truncateNode) Prefixes([]keys.Walkable) {}

func (n *truncateNode) Kind() string {
	return "truncateNode"
}

func (n *truncateNode) Init() error {
	n.isDone = false
	n.currentValue = n.documentMapping.NewDoc()
	return nil
}

func (n *truncateNode) Start() error { return nil }

func (n *truncateNode) Close() error { return nil }

func (n *truncateNode) Source() planNode { return nil }

func (n *truncateNode) Next() (bool, error) {
	if n.isDone {
		return false, nil
	}

	truncateOpts := options.WithIdentity(options.TruncateCollection(), n.p.identity)
	if n.filter.HasValue() {
		truncateOpts.SetFilter(n.filter.Value())
	}
	truncateOpts.SetPruneHistory(n.pruneHistory)
	if err := n.collection.Truncate(n.p.ctx, truncateOpts); err != nil {
		return false, err
	}

	n.currentValue.Fields[0] = true
	n.isDone = true
	return true, nil
}

func (n *truncateNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		var filter any
		if n.filter.HasValue() {
			filter = n.filter.Value()
		}
		return map[string]any{
			filterLabel:                 filter,
			request.PruneHistoryArgName: n.pruneHistory,
		}, nil
	case request.ExecuteExplain:
		return map[string]any{"executed": n.isDone}, nil
	default:
		return nil, ErrUnknownExplainRequestType
	}
}

func (p *Planner) Truncate(parsed *mapper.Mutation) (planNode, error) {
	col, err := p.db.GetCollectionByName(
		p.ctx,
		parsed.CollectionName,
		options.WithIdentity(options.GetCollectionByName(), p.identity),
	)
	if err != nil {
		return nil, err
	}

	return &truncateNode{
		p:            p,
		collection:   col,
		filter:       parsed.TruncateFilter,
		pruneHistory: parsed.PruneHistory,
		docMapper:    docMapper{parsed.DocumentMapping},
	}, nil
}
