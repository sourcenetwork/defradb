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
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// addNode is used to construct and execute
// an object add mutation.
//
// Add nodes are the simplest of the object mutations
// Each iteration of the plan, adds and returns one
// document, until we've exhausted the payload. No filtering
// or Select plans
type addNode struct {
	documentIterator
	docMapper

	p *Planner

	// cache information about the original data source
	// collection name, meta-data, etc.
	collection client.Collection

	// input map of fields and values
	input []map[string]any
	docs  []*client.Document

	didAdd bool

	results planNode

	execInfo addExecInfo

	addOptions []options.Enumerable[options.AddDocumentOptions]
}

type addExecInfo struct {
	// Total number of times addNode was executed.
	iterations uint64
}

func (n *addNode) Kind() string { return "addNode" }

func (n *addNode) Init() error { return nil }

func (n *addNode) docIDsToPrefixes(ids []string, desc client.CollectionVersion) ([]keys.Walkable, error) {
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

func documentsToDocIDs(docs ...*client.Document) []string {
	docIDs := make([]string, len(docs))
	for i, doc := range docs {
		docIDs[i] = doc.ID().String()
	}
	return docIDs
}

func (n *addNode) Start() error {
	n.docs = make([]*client.Document, len(n.input))

	for i, input := range n.input {
		doc, err := client.NewDocFromMap(n.p.ctx, input, n.collection.Version())
		if err != nil {
			return err
		}
		n.docs[i] = doc
	}

	return nil
}

func (n *addNode) Next() (bool, error) {
	n.execInfo.iterations++

	if !n.didAdd {
		err := n.collection.AddManyDocuments(n.p.ctx, n.docs, n.addOptions...)
		if err != nil {
			return false, err
		}

		prefixes, err := n.docIDsToPrefixes(documentsToDocIDs(n.docs...), n.collection.Version())
		if err != nil {
			return false, err
		}

		n.results.Prefixes(prefixes)

		err = n.results.Init()
		if err != nil {
			return false, err
		}

		err = n.results.Start()
		if err != nil {
			return false, err
		}
		n.didAdd = true
	}

	next, err := n.results.Next()
	n.currentValue = n.results.Value()
	return next, err
}

func (n *addNode) Prefixes(prefixes []keys.Walkable) { /* no-op */ }

func (n *addNode) Close() error {
	return n.results.Close()
}

func (n *addNode) Source() planNode { return n.results }

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *addNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return map[string]any{
			inputLabel: n.input,
		}, nil

	case request.ExecuteExplain:
		return map[string]any{
			"iterations": n.execInfo.iterations,
		}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}

func (p *Planner) AddDocs(parsed *mapper.Mutation) (planNode, error) {
	results, err := p.Select(&parsed.Select)
	if err != nil {
		return nil, err
	}

	add := &addNode{
		p:         p,
		input:     parsed.AddInput,
		results:   results,
		docMapper: docMapper{parsed.DocumentMapping},
		addOptions: []options.Enumerable[options.AddDocumentOptions]{
			options.WithIdentity(
				options.AddDocument().
					SetEncryptDoc(parsed.Encrypt).
					SetEncryptedFields(parsed.EncryptFields),
				p.identity,
			),
		},
	}

	col, err := p.db.GetCollectionByName(
		p.ctx,
		parsed.Name,
		options.WithIdentity(options.GetCollectionByName(), p.identity),
	)
	if err != nil {
		return nil, err
	}
	add.collection = col
	return add, nil
}
