// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"context"

	"github.com/sourcenetwork/defradb/datastore"
)

// IndexFieldDescription describes how a field is being indexed.
type IndexedFieldDescription struct {
	// Name contains the name of the field.
	Name string
	// Descending indicates whether the field is indexed in descending order.
	Descending bool
}

// IndexDescription describes an index.
type IndexDescription struct {
	// Name contains the name of the index.
	Name string
	// ID is the local identifier of this index.
	ID uint32
	// Fields contains the fields that are being indexed.
	Fields []IndexedFieldDescription
	// Unique indicates whether the index is unique.
	Unique bool
}

// IndexCreateRequestDescription describes an index creation request.
// It does not contain the ID, as it is not a valid field for the request body.
// Instead it should be automatically generated.
type IndexCreateRequestDescription struct {
	// Name contains the name of the index.
	Name string
	// Fields contains the fields that are being indexed.
	Fields []IndexedFieldDescription
	// Unique indicates whether the index is unique.
	Unique bool
}

// CollectionIndex is an interface for indexing documents in a collection.
type CollectionIndex interface {
	// Save indexes a document by storing indexed field values.
	// It doesn't retire previous values. For this [Update] should be used.
	Save(context.Context, datastore.Txn, *Document) error
	// Update updates an existing document in the index.
	// It removes the previous indexed field values and stores the new ones.
	Update(context.Context, datastore.Txn, *Document, *Document) error
	// Delete deletes an existing document from the index
	Delete(context.Context, datastore.Txn, *Document) error
	// Name returns the name of the index
	Name() string
	// Description returns the description of the index
	Description() IndexDescription
}

// CollectIndexedFields returns all fields that are indexed by all collection indexes.
func (d CollectionDefinition) CollectIndexedFields() []FieldDefinition {
	fieldsMap := make(map[string]bool)
	fields := make([]FieldDefinition, 0, len(d.Description.Indexes))
	for _, index := range d.Description.Indexes {
		for _, field := range index.Fields {
			if fieldsMap[field.Name] {
				// If the FieldDescription has already been added to the result do not add it a second time
				// this can happen if a field is referenced by multiple indexes
				continue
			}
			colField, ok := d.GetFieldByName(field.Name)
			if ok {
				fields = append(fields, colField)
			}
		}
	}
	return fields
}

// GetIndexesOnField returns all indexes that are indexing the given field.
// If the field is not the first field of a composite index, the index is not returned.
func (d CollectionDescription) GetIndexesOnField(fieldName string) []IndexDescription {
	result := []IndexDescription{}
	for _, index := range d.Indexes {
		if index.Fields[0].Name == fieldName {
			result = append(result, index)
		}
	}
	return result
}
