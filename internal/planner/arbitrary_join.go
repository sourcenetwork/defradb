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
	"fmt"
	"strings"

	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// A data-source that may yield child items, parent items, or both depending on configuration
type dataSource struct {
	pipeNode *pipeNode

	parentSource planNode
	childSource  planNode

	childIndex int

	lastParentDocIndex int
	lastChildDocIndex  int
}

func newDataSource(childIndex int) *dataSource {
	return &dataSource{
		childIndex:         childIndex,
		lastParentDocIndex: -1,
		lastChildDocIndex:  -1,
	}
}

func (n *dataSource) Init() error {
	// A docIndex of minus -1 indicated that nothing has been read yet
	n.lastChildDocIndex = -1
	n.lastParentDocIndex = -1

	if n.parentSource != nil {
		err := n.parentSource.Init()
		if err != nil {
			return err
		}
	}

	if n.childSource != nil {
		err := n.childSource.Init()
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *dataSource) Start() error {
	if n.parentSource != nil {
		err := n.parentSource.Start()
		if err != nil {
			return err
		}
	}

	if n.childSource != nil {
		err := n.childSource.Start()
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *dataSource) Spans(spans []core.Span) {
	if n.parentSource != nil {
		n.parentSource.Spans(spans)
	}

	if n.childSource != nil {
		n.childSource.Spans(spans)
	}
}

func (n *dataSource) Close() error {
	var err error
	if n.parentSource != nil {
		err = n.parentSource.Close()
		if err != nil {
			return err
		}
	}

	if n.childSource != nil {
		err = n.childSource.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *dataSource) Source() planNode {
	return n.parentSource
}

func (source *dataSource) mergeParent(
	keyFields []mapper.Field,
	destination *orderedMap,
	childIndexes []int,
) (bool, error) {
	// This needs to be set manually for each item, in case other nodes
	// aggregate items from the pipe progressing the docIndex beyond the first item
	// for example, if the child is sorted.
	source.pipeNode.docIndex = source.lastParentDocIndex
	defer func() {
		source.lastParentDocIndex = source.pipeNode.docIndex
	}()

	hasNext, err := source.parentSource.Next()
	if err != nil {
		return false, err
	}
	if !hasNext {
		return false, nil
	}

	value := source.parentSource.Value()
	key := generateKey(value, keyFields)

	destination.mergeParent(key, childIndexes, value)

	return true, nil
}

func (source *dataSource) appendChild(
	keyFields []mapper.Field,
	valuesByKey *orderedMap,
	mapping *core.DocumentMapping,
) (bool, error) {
	// Most of the time this will be the same document as the parent (with different rendering),
	// however if the child group is sorted it will be different, the child may also be missing
	// if it is filtered out by a child filter.  The parent will always exist, but may be
	// processed after the child if inner sorts shift the order.
	source.pipeNode.docIndex = source.lastChildDocIndex
	defer func() {
		source.lastChildDocIndex = source.pipeNode.docIndex
	}()

	hasNext, err := source.childSource.Next()
	if err != nil {
		return false, err
	}
	if !hasNext {
		return false, nil
	}

	// Note that even if the source yields both parent and child items, they may not be yielded in
	// the same order - we need to treat it as a new item, regenerating the key and potentially caching
	// it without yet receiving the parent-level details
	value := source.childSource.Value()
	key := generateKey(value, keyFields)

	valuesByKey.appendChild(key, source.childIndex, value, mapping)

	return true, nil
}

func join(
	sources []*dataSource,
	keyFields []mapper.Field,
	mapping *core.DocumentMapping,
) (*orderedMap, error) {
	result := orderedMap{
		values:       []core.Doc{},
		indexesByKey: map[string]int{},
	}

	childIndexes := make([]int, len(sources))
	for i, source := range sources {
		childIndexes[i] = source.childIndex
	}

	for _, source := range sources {
		var err error
		hasNextParent := source.parentSource != nil
		hasNextChild := source.childSource != nil

		for hasNextParent || hasNextChild {
			if hasNextParent {
				hasNextParent, err = source.mergeParent(keyFields, &result, childIndexes)
				if err != nil {
					return nil, err
				}
			}

			if hasNextChild {
				hasNextChild, err = source.appendChild(keyFields, &result, mapping)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return &result, nil
}

func generateKey(doc core.Doc, keyFields []mapper.Field) string {
	keyBuilder := strings.Builder{}
	for _, keyField := range keyFields {
		keyBuilder.WriteString(fmt.Sprint(keyField.Index))
		keyBuilder.WriteString(fmt.Sprintf("_%v_", doc.Fields[keyField.Index]))
	}
	return keyBuilder.String()
}

// A specialized collection that allows retrieval of items by key whilst preserving the order
// in which they were added.
type orderedMap struct {
	values       []core.Doc
	indexesByKey map[string]int
}

func (m *orderedMap) mergeParent(key string, childIndexes []int, value core.Doc) {
	index, exists := m.indexesByKey[key]
	if exists {
		existingValue := m.values[index]

		// copy every value from the child, apart from the child-indexes
	propertyLoop:
		for cellIndex, cellValue := range value.Fields {
			for _, childIndex := range childIndexes {
				if cellIndex == childIndex {
					continue propertyLoop
				}
			}
			existingValue.Fields[cellIndex] = cellValue
		}

		return
	}

	// If the value is new, we can safely set the child group to an empty
	// collection (required if children are filtered out)
	for _, childAddress := range childIndexes {
		// the parent may have come from a pipe using a smaller doc mapping,
		// if so we need to extend the field slice.
		if childAddress >= len(value.Fields) {
			newFields := make(core.DocFields, childAddress+1)
			copy(newFields, value.Fields)
			value.Fields = newFields
		}
		value.Fields[childAddress] = []core.Doc{}
	}

	index = len(m.values)
	m.values = append(m.values, value)
	m.indexesByKey[key] = index
}

func (m *orderedMap) appendChild(key string, childIndex int, value core.Doc, mapping *core.DocumentMapping) {
	index, exists := m.indexesByKey[key]
	var parent core.Doc
	if !exists {
		index = len(m.values)

		parent = mapping.NewDoc()
		m.values = append(m.values, parent)

		m.indexesByKey[key] = index
	} else {
		parent = m.values[index]
	}

	childProperty := parent.Fields[childIndex]
	if childProperty == nil {
		childProperty = []core.Doc{
			value,
		}
		parent.Fields[childIndex] = childProperty
		return
	}

	childCollection := childProperty.([]core.Doc)
	parent.Fields[childIndex] = append(childCollection, value)
}
