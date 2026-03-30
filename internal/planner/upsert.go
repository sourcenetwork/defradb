// Copyright 2024 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

type upsertNode struct {
	documentIterator
	docMapper

	p             *Planner
	collection    client.Collection
	filter        *mapper.Filter
	addInput      map[string]any
	updateInput   map[string]any
	isInitialized bool
	source        planNode
	origScanNode  *scanNode
	valuesNode    *valuesNode
}

// Next only returns once.
func (n *upsertNode) Next() (bool, error) {
	if !n.isInitialized {
		var updater bool
		next, err := n.source.Next()
		if err != nil {
			return false, err
		}
		if next {
			n.currentValue = n.source.Value()
			// make sure multiple documents do not match
			next, err := n.source.Next()
			if err != nil {
				return false, err
			}
			if next {
				return false, ErrUpsertMultipleDocuments
			}
			docID, err := client.NewDocIDFromString(n.currentValue.GetID())
			if err != nil {
				return false, err
			}
			getOpts := options.WithIdentity(options.GetDocument(), n.p.identity)
			doc, err := n.collection.GetDocument(n.p.ctx, docID, getOpts)
			if err != nil {
				return false, err
			}
			for k, v := range n.updateInput {
				if err := doc.Set(n.p.ctx, k, v); err != nil {
					return false, NewErrSetDocField(err, k)
				}
			}
			updateOpts := options.WithIdentity(options.UpdateDocument(), n.p.identity)
			err = n.collection.UpdateDocument(n.p.ctx, doc, updateOpts)
			if err != nil {
				return false, err
			}
			coreDoc, err := core.DocFromClient(doc, n.documentMapping)
			if err != nil {
				return false, err
			}

			n.valuesNode.docs.AddDoc(coreDoc)

			updater = true
		} else {
			doc, err := client.NewDocFromMap(n.p.ctx, n.addInput, n.collection.Version())
			if err != nil {
				return false, err
			}
			addOpts := options.WithIdentity(options.AddDocument(), n.p.identity)
			err = n.collection.AddDocument(n.p.ctx, doc, addOpts)
			if err != nil {
				return false, err
			}

			prefixes, err := n.docIDsToPrefixes(documentsToDocIDs(doc), n.collection.Version())
			if err != nil {
				return false, err
			}

			n.source.Prefixes(prefixes)
		}

		if updater {
			// we have cached the document result set from the original Select
			// in the valuesNode, now we can replace the original scanNode with
			// our valuesNode, and avoid any additional fetches/kv ops.
			// This is cheaper than building two seperate plans.
			err := n.p.walkAndReplacePlan(n.source, n.origScanNode, n.valuesNode)
			if err != nil {
				return false, err
			}
			// The original scanNode is now orphaned (replaced by valuesNode in the plan tree).
			// Close it to release its fetcher's iterator, otherwise it leaks.
			if err := n.origScanNode.Close(); err != nil {
				return false, err
			}
		}

		err = n.source.Init()
		if err != nil {
			return false, err
		}
		n.isInitialized = true
	}
	next, err := n.source.Next()
	if err != nil {
		return false, err
	}
	if !next {
		return false, nil
	}
	n.currentValue = n.source.Value()
	return true, nil
}

func (n *upsertNode) docIDsToPrefixes(ids []string, desc client.CollectionVersion) ([]keys.Walkable, error) {
	shortID, err := id.GetShortCollectionID(n.p.ctx, desc.CollectionID)
	if err != nil {
		return nil, err
	}

	prefixes := make([]keys.Walkable, len(ids))
	for i, id := range ids {
		prefixes[i] = keys.DataStoreKey{
			CollectionShortID: shortID,
			DocID:             id,
		}
	}
	return prefixes, nil
}

func (n *upsertNode) Kind() string {
	return "upsertNode"
}

func (n *upsertNode) Prefixes(prefixes []keys.Walkable) {
	n.source.Prefixes(prefixes)
}

func (n *upsertNode) Init() error {
	err := n.source.Init()
	if err != nil {
		return err
	}

	n.origScanNode = getNode[*scanNode](n.source)
	return nil
}

func (n *upsertNode) Start() error {
	return n.source.Start()
}

func (n *upsertNode) Close() error {
	return n.source.Close()
}

func (n *upsertNode) Source() planNode {
	return n.source
}

func (n *upsertNode) simpleExplain() (map[string]any, error) {
	simpleExplainMap := map[string]any{}

	// Add the filter attribute
	simpleExplainMap[filterLabel] = n.filter.ToMap(n.documentMapping)

	// Add the attribute that represents the values to add or update.
	simpleExplainMap[updateInputLabel] = n.updateInput
	simpleExplainMap[addInputLabel] = n.addInput

	return simpleExplainMap, nil
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *upsertNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return n.simpleExplain()

	case request.ExecuteExplain:
		return map[string]any{}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}

func (p *Planner) UpsertDocs(parsed *mapper.Mutation) (planNode, error) {
	upsert := &upsertNode{
		p:           p,
		filter:      parsed.Filter,
		updateInput: parsed.UpdateInput,
		docMapper:   docMapper{parsed.DocumentMapping},
	}

	if len(parsed.AddInput) > 0 {
		upsert.addInput = parsed.AddInput[0]
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
	upsert.collection = col

	// create the results Select node
	resultsNode, err := p.Select(&parsed.Select)
	if err != nil {
		return nil, err
	}
	upsert.source = resultsNode
	upsert.valuesNode = p.newContainerValuesNode(nil)

	return upsert, nil
}
