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

	"github.com/sourcenetwork/defradb/core"
)

// A data-source that may yield child items, parent items, or both depending on configuration
type dataSource struct {
	pipeNode *pipeNode

	parentSource planNode
	childSource  planNode

	childName string

	lastParentDocIndex int
	lastChildDocIndex  int
}

func newDataSource(childName string) dataSource {
	return dataSource{
		childName:          childName,
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

func (n *dataSource) Spans(spans core.Spans) {
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

func (source *dataSource) mergeParent(keyFields []string, destination *orderedMap) (map[string]interface{}, bool, error) {
	// This needs to be set manually for each item, in case other nodes
	// aggregate items from the pipe progressing the docIndex beyond the first item
	// for example, if the child is sorted.
	source.pipeNode.docIndex = source.lastParentDocIndex
	defer func() {
		source.lastParentDocIndex = source.pipeNode.docIndex
	}()

	hasNext, err := source.parentSource.Next()
	if err != nil {
		return nil, false, err
	}
	if !hasNext {
		return nil, false, nil
	}

	value := source.parentSource.Values()
	key := generateKey(value, keyFields)

	destination.mergeParent(key, source.childName, value)

	return value, true, nil
}

func (source *dataSource) appendChild(keyFields []string, valuesByKey *orderedMap) (map[string]interface{}, bool, error) {
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
		return nil, false, err
	}
	if !hasNext {
		return nil, false, nil
	}

	// Note that even if the source yields both parent and child items, they may not be yielded in
	// the same order - we need to treat it as a new item, regenerating the key and potentially caching
	// it without yet receiving the parent-level details
	value := source.childSource.Values()
	key := generateKey(value, keyFields)

	valuesByKey.appendChild(key, source.childName, value)

	return value, true, nil
}

func join(sources []dataSource, keyFields []string) (*orderedMap, error) {
	result := orderedMap{
		values:       []map[string]interface{}{},
		indexesByKey: map[string]int{},
	}

	for _, source := range sources {
		var err error
		hasNextParent := source.parentSource != nil
		hasNextChild := source.childSource != nil
		for hasNextParent || hasNextChild {
			if hasNextParent {
				_, hasNextParent, err = source.mergeParent(keyFields, &result)
				if err != nil {
					return nil, err
				}
			}

			if hasNextChild {
				_, hasNextChild, err = source.appendChild(keyFields, &result)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return &result, nil
}

func generateKey(doc map[string]interface{}, keyFields []string) string {
	keyBuilder := strings.Builder{}
	for _, keyField := range keyFields {
		keyBuilder.WriteString(keyField)
		keyBuilder.WriteString(fmt.Sprintf("%v", doc[keyField]))
	}
	return keyBuilder.String()
}

// A specialized collection that allows retrieval of items by key whilst preserving the order
// in which they were added.
type orderedMap struct {
	values       []map[string]interface{}
	indexesByKey map[string]int
}

func (m *orderedMap) mergeParent(key string, childAddress string, value map[string]interface{}) {
	index, exists := m.indexesByKey[key]
	if exists {
		existingValue := m.values[index]
		for property, cellValue := range value {
			if property == childAddress {
				continue
			}
			existingValue[property] = cellValue
		}
		return
	}

	// If the value is new, we can safely set the child group to an empty collection (required if children are filtered out)
	value[childAddress] = []map[string]interface{}{}

	index = len(m.values)
	m.values = append(m.values, value)
	m.indexesByKey[key] = index
}

func (m *orderedMap) appendChild(key string, childAddress string, value map[string]interface{}) {
	index, exists := m.indexesByKey[key]
	var parent map[string]interface{}
	if !exists {
		index = len(m.values)

		parent = map[string]interface{}{}
		m.values = append(m.values, parent)

		m.indexesByKey[key] = index
	} else {
		parent = m.values[index]
	}

	childProperty, hasChildCollection := parent[childAddress]
	if !hasChildCollection {
		childProperty = []map[string]interface{}{
			value,
		}
		parent[childAddress] = childProperty
		return
	}

	childCollection := childProperty.([]map[string]interface{})
	parent[childAddress] = append(childCollection, value)
}
