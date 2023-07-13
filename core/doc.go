// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package core provides commonly shared interfaces and building blocks.
*/
package core

import (
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

// DocKeyFieldIndex is the index of the key field in a document.
const DocKeyFieldIndex int = 0

// DocFields is a slice of fields in a document.
type DocFields []any

// Doc is a document.
type Doc struct {
	// If true, this Doc will not be rendered, but will still be passed through
	// the plan graph just like any other document.
	Hidden bool

	Fields DocFields
	Status client.DocumentStatus
	// The id of the schema version that this document is currently at.  This includes
	// any migrations that may have been run.
	SchemaVersionID string
}

// GetKey returns the DocKey for this document.
//
// Will panic if the document is empty.
func (d *Doc) GetKey() string {
	key, _ := d.Fields[DocKeyFieldIndex].(string)
	return key
}

// SetKey sets the DocKey for this document.
//
// Will panic if the document has not been initialised with fields.
func (d *Doc) SetKey(key string) {
	d.Fields[DocKeyFieldIndex] = key
}

// Clone returns a deep copy of this document.
func (d *Doc) Clone() Doc {
	cp := Doc{
		Fields: make(DocFields, len(d.Fields)),
	}

	for i, v := range d.Fields {
		switch typedFieldValue := v.(type) {
		case Doc:
			cp.Fields[i] = typedFieldValue.Clone()
		case []Doc:
			innerMaps := make([]Doc, len(typedFieldValue))
			for j, innerDoc := range typedFieldValue {
				innerMaps[j] = innerDoc.Clone()
			}
			cp.Fields[i] = innerMaps
		default:
			cp.Fields[i] = v
		}
	}

	return cp
}

// RenderKey is a key that should be rendered into the document.
type RenderKey struct {
	// The field index to be rendered.
	Index int

	// The key by which the field contents should be rendered into.
	Key string
}

type mappingTypeInfo struct {
	// The index at which the type name is to be held
	Index int

	// The name of the host type
	Name string
}

// DocumentMapping is a mapping of a document.
type DocumentMapping struct {
	// The type information for the object, if provided.
	typeInfo immutable.Option[mappingTypeInfo]

	// The set of fields that should be rendered.
	//
	// Fields not in this collection will not be rendered to the consumer.
	// Collection-item indexes do not have to pair up with field indexes and
	// items should not be accessed this way.
	RenderKeys []RenderKey

	// The set of fields available using this mapping.
	//
	// If a field-name is not in this collection, it essentially doesn't exist.
	// Collection should include fields that are not rendered to the consumer.
	// Multiple fields may exist for any given name (for example if a property
	// exists under different aliases/filters).
	IndexesByName map[string][]int

	// The next index available for use.
	//
	// Also useful for identifying how many fields a document should have.
	nextIndex int

	// The collection of child mappings for this object.
	//
	// Indexes correspond exactly to field indexes, however entries may be default
	// if the field is unmappable (e.g. integer fields).
	ChildMappings []*DocumentMapping
}

// NewDocumentMapping instantiates a new DocumentMapping instance.
func NewDocumentMapping() *DocumentMapping {
	return &DocumentMapping{
		IndexesByName: map[string][]int{},
	}
}

// CloneWithoutRender deep copies the source mapping skipping over the RenderKeys.
func (source *DocumentMapping) CloneWithoutRender() *DocumentMapping {
	result := DocumentMapping{
		typeInfo:      source.typeInfo,
		IndexesByName: make(map[string][]int, len(source.IndexesByName)),
		nextIndex:     source.nextIndex,
		ChildMappings: make([]*DocumentMapping, len(source.ChildMappings)),
	}

	for externalName, sourceIndexes := range source.IndexesByName {
		indexes := make([]int, len(sourceIndexes))
		copy(indexes, sourceIndexes)
		result.IndexesByName[externalName] = indexes
	}

	for i, childMapping := range source.ChildMappings {
		if childMapping != nil {
			result.ChildMappings[i] = childMapping.CloneWithoutRender()
		}
	}

	return &result
}

// GetNextIndex returns the next index available for use.
//
// Also useful for identifying how many fields a document should have.
func (mapping *DocumentMapping) GetNextIndex() int {
	return mapping.nextIndex
}

// NewDoc instantiates a new Doc from this mapping, ensuring that the Fields
// collection is constructed with the required length/indexes.
func (mapping *DocumentMapping) NewDoc() Doc {
	return Doc{
		Fields: make(DocFields, mapping.nextIndex),
	}
}

// SetFirstOfName overwrites the first field of this name with the given value.
//
// Will panic if the field does not exist.
func (mapping *DocumentMapping) SetFirstOfName(d *Doc, name string, value any) {
	d.Fields[mapping.IndexesByName[name][0]] = value
}

// FirstOfName returns the value of the first field of the given name.
//
// Will panic if the field does not exist (but not if it's value is default).
func (mapping *DocumentMapping) FirstOfName(d Doc, name string) any {
	return d.Fields[mapping.FirstIndexOfName(name)]
}

// FirstIndexOfName returns the first field index of the given name.
//
// Will panic if the field does not exist.
func (mapping *DocumentMapping) FirstIndexOfName(name string) int {
	return mapping.IndexesByName[name][0]
}

// ToMap renders the given document to map[string]any format using
// the given mapping.
//
// Will not return fields without a render key, or any child documents
// marked as Hidden.
func (mapping *DocumentMapping) ToMap(doc Doc) map[string]any {
	mappedDoc := make(map[string]any, len(mapping.RenderKeys))
	for _, renderKey := range mapping.RenderKeys {
		value := doc.Fields[renderKey.Index]
		var renderValue any
		switch innerV := value.(type) {
		case []Doc:
			innerMapping := mapping.ChildMappings[renderKey.Index]
			innerArray := []map[string]any{}
			for _, innerDoc := range innerV {
				if innerDoc.Hidden {
					continue
				}
				innerArray = append(innerArray, innerMapping.ToMap(innerDoc))
			}
			renderValue = innerArray
		case Doc:
			innerMapping := mapping.ChildMappings[renderKey.Index]
			renderValue = innerMapping.ToMap(innerV)
		default:
			if mapping.typeInfo.HasValue() && renderKey.Index == mapping.typeInfo.Value().Index {
				renderValue = mapping.typeInfo.Value().Name
			} else {
				renderValue = innerV
			}
		}
		mappedDoc[renderKey.Key] = renderValue
	}
	return mappedDoc
}

// Add appends the given index and name to the mapping.
func (mapping *DocumentMapping) Add(index int, name string) {
	inner := mapping.IndexesByName[name]
	inner = append(inner, index)
	mapping.IndexesByName[name] = inner

	if index >= mapping.nextIndex {
		mapping.nextIndex = index + 1
	}
}

// SetTypeName sets the type name for this mapping.
func (mapping *DocumentMapping) SetTypeName(typeName string) {
	index := mapping.GetNextIndex()
	mapping.Add(index, request.TypeNameFieldName)
	mapping.typeInfo = immutable.Some(mappingTypeInfo{
		Index: index,
		Name:  typeName,
	})
}

// SetChildAt sets the given child mapping at the given index.
//
// If the index is greater than the ChildMappings length the collection will
// grow.
func (m *DocumentMapping) SetChildAt(index int, childMapping *DocumentMapping) {
	var newMappings []*DocumentMapping
	if index >= len(m.ChildMappings)-1 {
		newMappings = make([]*DocumentMapping, index+1)
		copy(newMappings, m.ChildMappings)
	} else {
		newMappings = m.ChildMappings
	}

	newMappings[index] = childMapping
	m.ChildMappings = newMappings
}

// TryToFindNameFromIndex returns the corresponding name of the given index.
//
// Additionally, will also return true if the index was found, and false otherwise.
func (mapping *DocumentMapping) TryToFindNameFromIndex(targetIndex int) (string, bool) {
	// Try to find the name of this index in the IndexesByName.
	for name, indexes := range mapping.IndexesByName {
		for _, index := range indexes {
			if index == targetIndex {
				return name, true
			}
		}
	}

	return "", false
}
