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
	"sort"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

type updateNode struct {
	documentIterator
	docMapper

	p *Planner

	collection client.Collection

	filter *mapper.Filter

	docIDs []string

	// input map of fields and values
	input map[string]any

	results planNode

	execInfo updateExecInfo
}

type updateExecInfo struct {
	// Total number of times updateNode was executed.
	iterations uint64

	// Total number of successful updates.
	updates uint64
}

// Next only returns once.
func (n *updateNode) Next() (bool, error) {
	n.execInfo.iterations++

	next, err := n.results.Next()
	if err != nil {
		return false, err
	}
	if !next {
		return false, nil
	}

	n.currentValue = n.results.Value()

	docID, err := client.NewDocIDFromString(n.currentValue.GetID())
	if err != nil {
		return false, err
	}
	getOpts := options.WithIdentity(options.GetDocument(), n.p.identity)
	doc, err := n.collection.GetDocument(n.p.ctx, docID, getOpts)
	if err != nil {
		return false, err
	}
	for k, v := range n.input {
		if err := doc.Set(n.p.ctx, k, v); err != nil {
			return false, err
		}
	}
	updateOpts := options.WithIdentity(options.UpdateDocument(), n.p.identity)
	err = n.collection.UpdateDocument(n.p.ctx, doc, updateOpts)
	if err != nil {
		return false, err
	}

	n.execInfo.updates++

	coreDoc, err := core.DocFromClient(doc, n.documentMapping)
	if err != nil {
		return false, err
	}

	n.currentValue = coreDoc
	return true, nil
}

func (n *updateNode) Kind() string { return "updateNode" }

func (n *updateNode) Prefixes(prefixes []keys.Walkable) { n.results.Prefixes(prefixes) }

func (n *updateNode) Init() error {
	return n.results.Init()
}

func (n *updateNode) Start() error {
	return n.results.Start()
}

func (n *updateNode) Close() error {
	return n.results.Close()
}

func (n *updateNode) Source() planNode { return n.results }

func (n *updateNode) simpleExplain() (map[string]any, error) {
	simpleExplainMap := map[string]any{}

	// Add the document id(s) that request wants to update.
	simpleExplainMap[request.DocIDArgName] = n.docIDs

	// Add the filter attribute if it exists, otherwise have it nil.
	if n.filter == nil {
		simpleExplainMap[filterLabel] = nil
	} else {
		simpleExplainMap[filterLabel] = n.filter.ToMap(n.documentMapping)
	}

	// Add the attribute that represents the patch to update with.
	simpleExplainMap[inputLabel] = n.input

	return simpleExplainMap, nil
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *updateNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return n.simpleExplain()

	case request.ExecuteExplain:
		return map[string]any{
			"iterations": n.execInfo.iterations,
			"updates":    n.execInfo.updates,
		}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}

func (p *Planner) UpdateDocs(parsed *mapper.Mutation) (planNode, error) {
	update := &updateNode{
		p:         p,
		filter:    parsed.Filter,
		docIDs:    parsed.DocIDs.Value(),
		input:     parsed.UpdateInput,
		docMapper: docMapper{parsed.DocumentMapping},
	}

	// get collection
	col, err := p.db.GetCollectionByName(
		p.ctx,
		parsed.Name,
		options.WithIdentity(options.GetCollectionByName(), p.identity),
	)
	if err != nil {
		return nil, err
	}
	update.collection = col

	// shallow and deep copy on the fields since we're going to mutate
	preUpdateSelect := parsed.Select
	preUpdateSelect.Fields = make([]mapper.Requestable, 0)

	selectFieldsToDelete := make([]int, 0)

	// Split fields between inner (pre-update filter/scan) and outer (post-update render) selects.
	// The inner select only needs base fields and filter-related relations (SkipResolve).
	// Render-only relations go only to the outer select to avoid unnecessary type joins.
	for i, field := range parsed.Select.Fields {
		if _, exists := request.ReservedFields[field.GetName()]; exists {
			continue
		}
		switch f := field.(type) {
		case *mapper.Select:
			if f.SkipResolve {
				// Filter-only relation: include in inner select, remove from outer
				preUpdateSelect.Fields = append(preUpdateSelect.Fields, field)
				selectFieldsToDelete = append(selectFieldsToDelete, i)
			}
		case *mapper.Field:
			preUpdateSelect.Fields = append(preUpdateSelect.Fields, field)
		}
	}

	// removed unnecessary fields we get from the pre update select
	parsed.Select.Fields = deleteIndexes(parsed.Select.Fields, selectFieldsToDelete)

	selectNode, err := p.Select(&preUpdateSelect)
	if err != nil {
		return nil, err
	}

	// Wire the inner selectTopNode's plan before it gets wrapped in the
	// outer plan tree. This is needed because expandTypeJoin only expands
	// the child side and won't reach this nested selectTopNode when a
	// relation sub-select creates a type join on the outer select.
	if top, ok := selectNode.(*selectTopNode); ok {
		top.planNode = top.selectNode
	}

	update.results = selectNode

	// The outer select only renders the post-update results. The inner select already
	// handles pre-update filtering (by filter and/or docIDs), so we clear both on the
	// outer select to prevent re-evaluation against potentially changed values.
	parsed.Select.Filter = nil
	parsed.Select.DocIDs = immutable.None[[]string]()
	return p.SelectFromSource(&parsed.Select, update, true, update.collection)
}

func deleteIndexes[T any](s []T, idx []int) []T {
	if len(idx) == 0 {
		return s
	}

	sort.Ints(idx)

	out := s[:0]
	j := 0

	for i := range s {
		if j < len(idx) && i == idx[j] {
			j++
			continue
		}
		out = append(out, s[i])
	}

	return out
}
