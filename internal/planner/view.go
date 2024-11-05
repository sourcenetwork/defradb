// Copyright 2023 Democratized Data Foundation
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
	"github.com/ipfs/go-datastore/query"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// viewNode processes queries to a Defra View constructed from a base query ahead of time.
type viewNode struct {
	docMapper

	p      *Planner
	desc   client.CollectionDescription
	source planNode

	// This is cached as a boolean to save rediscovering this in the main Next/Value iteration loop
	hasTransform bool
}

func (p *Planner) View(query *mapper.Select, col client.Collection) (planNode, error) {
	// For now, we assume a single source.  This will need to change if/when we support multiple sources
	querySource := (col.Description().Sources[0].(*client.QuerySource))
	hasTransform := querySource.Transform.HasValue()

	var source planNode
	if col.Description().IsMaterialized {
		source = p.newCachedViewFetcher(col.Definition(), query.DocumentMapping)
	} else {
		m, err := mapper.ToSelect(p.ctx, p.db, mapper.ObjectSelection, &querySource.Query)
		if err != nil {
			return nil, err
		}

		source, err = p.Select(m)
		if err != nil {
			return nil, err
		}

		if hasTransform {
			source = p.Lens(source, query.DocumentMapping, col)
		}
	}

	viewNode := &viewNode{
		p:            p,
		desc:         col.Description(),
		source:       source,
		docMapper:    docMapper{query.DocumentMapping},
		hasTransform: hasTransform,
	}

	return viewNode, nil
}

func (n *viewNode) Init() error {
	return n.source.Init()
}

func (n *viewNode) Start() error {
	return n.source.Start()
}

func (n *viewNode) Spans(spans []core.Span) {
	n.source.Spans(spans)
}

func (n *viewNode) Next() (bool, error) {
	return n.source.Next()
}

func (n *viewNode) Value() core.Doc {
	// The source mapping will differ from this node's (request) mapping if either a Lens transform is
	// involved, if the the view is materialized, or if any kind of operation is performed on the result
	// of the query (such as a filter or aggregate in the user-request), so we must convert the returned
	// documents to the request mapping
	return convertBetweenMaps(n.source.DocumentMap(), n.documentMapping, n.source.Value())
}

func (n *viewNode) Source() planNode {
	return n.source
}

func (n *viewNode) Kind() string {
	return "viewNode"
}

func (n *viewNode) Close() error {
	if n.source != nil {
		err := n.source.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func convertBetweenMaps(srcMap *core.DocumentMapping, dstMap *core.DocumentMapping, src core.Doc) core.Doc {
	dst := dstMap.NewDoc()

	srcRenderKeysByIndex := map[int]string{}
	for _, renderKey := range srcMap.RenderKeys {
		srcRenderKeysByIndex[renderKey.Index] = renderKey.Key
	}

	for underlyingName, srcIndexes := range srcMap.IndexesByName {
		for _, srcIndex := range srcIndexes {
			if srcIndex >= len(src.Fields) {
				// Several system fields are not included in schema only types, and there is a mismatch somewhere
				// that means we have to handle them here with a continue
				continue
			}

			var dstName string
			if key, ok := srcRenderKeysByIndex[srcIndex]; ok {
				dstName = key
			} else {
				dstName = underlyingName
			}

			dstIndexes, dstHasField := dstMap.IndexesByName[dstName]
			if !dstHasField {
				continue
			}

			for _, dstIndex := range dstIndexes {
				var srcValue any
				if srcIndex < len(srcMap.ChildMappings) && srcMap.ChildMappings[srcIndex] != nil {
					if dstIndex >= len(dstMap.ChildMappings) || dstMap.ChildMappings[dstIndex] == nil {
						continue
					}

					switch inner := src.Fields[srcIndex].(type) {
					case core.Doc:
						srcValue = convertBetweenMaps(srcMap.ChildMappings[srcIndex], dstMap.ChildMappings[dstIndex], inner)

					case []core.Doc:
						dstInners := make([]core.Doc, len(inner))
						for i, srcInnerDoc := range inner {
							dstInners[i] = convertBetweenMaps(srcMap.ChildMappings[srcIndex], dstMap.ChildMappings[dstIndex], srcInnerDoc)
						}
						srcValue = dstInners
					}
				} else {
					srcValue = src.Fields[srcIndex]
				}

				dst.Fields[dstIndex] = srcValue
			}
		}
	}

	return dst
}

// cachedViewFetcher is a planner node that fetches view items from a materialized cache.
type cachedViewFetcher struct {
	docMapper
	documentIterator

	def client.CollectionDefinition
	p   *Planner

	queryResults query.Results
}

var _ planNode = (*cachedViewFetcher)(nil)

func (p *Planner) newCachedViewFetcher(
	def client.CollectionDefinition,
	mapping *core.DocumentMapping,
) *cachedViewFetcher {
	return &cachedViewFetcher{
		def:       def,
		p:         p,
		docMapper: docMapper{mapping},
	}
}

func (n *cachedViewFetcher) Init() error {
	if n.queryResults != nil {
		err := n.queryResults.Close()
		if err != nil {
			return err
		}
		n.queryResults = nil
	}

	prefix := keys.NewViewCacheColPrefix(n.def.Description.RootID)

	var err error
	n.queryResults, err = n.p.txn.Datastore().Query(n.p.ctx, query.Query{
		Prefix: prefix.ToString(),
	})
	if err != nil {
		return err
	}

	return nil
}

func (n *cachedViewFetcher) Start() error {
	return nil
}

func (n *cachedViewFetcher) Spans(spans []core.Span) {
	// no-op
}

func (n *cachedViewFetcher) Next() (bool, error) {
	result, hasNext := n.queryResults.NextSync()
	if !hasNext || result.Error != nil {
		return false, result.Error
	}

	var err error
	n.currentValue, err = core.UnmarshalViewItem(n.documentMapping, result.Value)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (n *cachedViewFetcher) Source() planNode {
	return nil
}

func (n *cachedViewFetcher) Kind() string {
	return "cachedViewFetcher"
}

func (n *cachedViewFetcher) Close() error {
	return n.queryResults.Close()
}
