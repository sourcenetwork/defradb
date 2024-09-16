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
	"github.com/sourcenetwork/immutable/enumerable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
)

// viewNode applies a lens transform to data yielded from the source node.
//
// It may return a different number of documents to that yielded by its source,
// and there is no guarentee that those documents will actually exist as documents
// in Defra (they may be created by the transform).
type lensNode struct {
	docMapper
	documentIterator

	p          *Planner
	source     planNode
	collection client.CollectionDescription

	input  enumerable.Queue[map[string]any]
	output enumerable.Enumerable[map[string]any]
}

func (p *Planner) Lens(source planNode, docMap *core.DocumentMapping, col client.Collection) *lensNode {
	return &lensNode{
		docMapper:  docMapper{docMap},
		p:          p,
		source:     source,
		collection: col.Description(),
	}
}

func (n *lensNode) Init() error {
	n.input = enumerable.NewQueue[map[string]any]()

	pipe, err := n.p.db.LensRegistry().MigrateUp(n.p.ctx, n.input, n.collection.ID)
	if err != nil {
		return err
	}

	n.output = pipe

	return n.source.Init()
}

func (n *lensNode) Start() error {
	return n.source.Start()
}

func (n *lensNode) Spans(spans core.Spans) {
	n.source.Spans(spans)
}

func (n *lensNode) Next() (bool, error) {
	hasNext, err := n.output.Next()
	if err != nil {
		return false, err
	}

	if hasNext {
		lensDoc, err := n.output.Value()
		if err != nil {
			return false, err
		}

		n.currentValue = n.toDoc(n.documentMapping, lensDoc)
		return true, nil
	}

	sourceHasNext, err := n.source.Next()
	if err != nil {
		return false, err
	}

	if !sourceHasNext {
		return false, nil
	}

	sourceDoc := n.source.Value()
	sourceLensDoc := n.source.Source().DocumentMap().ToMap(sourceDoc)

	err = n.input.Put(sourceLensDoc)
	if err != nil {
		return false, err
	}

	return n.Next()
}

func (n *lensNode) toDoc(mapping *core.DocumentMapping, mapDoc map[string]any) core.Doc {
	status := client.Active
	properties := make([]any, len(mapDoc))

	for fieldName, fieldValue := range mapDoc {
		if fieldName == request.DocIDFieldName && fieldValue != nil {
			properties[core.DocIDFieldIndex] = fieldValue.(string)
			continue
		}

		if fieldName == request.DeletedFieldName {
			if wasDeleted, ok := fieldValue.(bool); ok {
				if wasDeleted {
					status = client.Deleted
				}
			}
			continue
		}

		indexes := mapping.IndexesByName[fieldName]
		if len(indexes) == 0 {
			// Note: This can happen if a migration returns a field that
			// we do not know about. In which case we have to skip it.
			continue
		}
		// Take the last index of this name, this is in order to be consistent with other
		// similar logic, for example when converting a core.Doc to a map before passing it
		// into a lens transform.
		fieldIndex := indexes[len(indexes)-1]

		if fieldIndex < len(mapping.ChildMappings) && mapping.ChildMappings[fieldIndex] != nil {
			switch typedValue := fieldValue.(type) {
			case map[string]any:
				fieldValue = n.toDoc(mapping.ChildMappings[fieldIndex], typedValue)

			case []any:
				values := make([]core.Doc, 0, len(typedValue))
				for _, val := range typedValue {
					innerDoc := n.toDoc(mapping.ChildMappings[fieldIndex], val.(map[string]any))
					values = append(values, innerDoc)
				}
				fieldValue = values
			}
		}

		if len(properties) <= fieldIndex {
			// Because the document is sourced from another mapping, we may still need to grow
			// the resultant field set. We cannot use [append] because the index of each field
			// must still correspond to it's field ID.
			originalProps := properties
			properties = make([]any, fieldIndex+1)
			copy(properties, originalProps)
		}
		properties[fieldIndex] = fieldValue
	}

	return core.Doc{
		Fields:          properties,
		SchemaVersionID: n.collection.SchemaVersionID,
		Status:          status,
	}
}

func (n *lensNode) Source() planNode {
	return n.source
}

func (n *lensNode) Kind() string {
	return "lensNode"
}

func (n *lensNode) Close() error {
	if n.source != nil {
		err := n.source.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
