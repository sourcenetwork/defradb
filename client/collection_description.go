// Copyright 2024 Democratized Data Foundation
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
	"encoding/json"
	"fmt"
	"math"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
)

// CollectionDescription with no known root will take this ID as their temporary RootID.
//
// Orphan CollectionDescriptions are typically created when setting migrations from schema versions
// that do not yet exist.  The OrphanRootID will be replaced with the actual RootID once a full chain
// of schema versions leading back to a schema version used by a collection with a non-orphan RootID
// has been established.
const OrphanRootID uint32 = math.MaxUint32

// CollectionDescription describes a Collection and all its associated metadata.
type CollectionDescription struct {
	// Name contains the name of the collection.
	//
	// It is conceptually local to the node hosting the DefraDB instance, but currently there
	// is no means to update the local value so that it differs from the (global) schema name.
	Name immutable.Option[string]

	// ID is the local identifier of this collection.
	//
	// It is immutable.
	ID uint32

	// RootID is the local root identifier of this collection, linking together a chain of
	// collection instances on different schema versions.
	//
	// Collections sharing the same RootID will be compatable with each other, with the documents
	// within them shared and yielded as if they were in the same set, using Lens transforms to
	// migrate between schema versions when provided.
	RootID uint32

	// The ID of the schema version that this collection is at.
	SchemaVersionID string

	// Sources is the set of sources from which this collection draws data.
	//
	// Currently supported source types are:
	// - [QuerySource]
	// - [CollectionSource]
	Sources []any

	// Fields contains the fields local to the node within this Collection.
	//
	// Most fields defined here will also be present on the [SchemaDescription]. A notable
	// exception to this are the fields of the (optional) secondary side of a relation
	// which are local only, and will not be present on the [SchemaDescription].
	Fields []CollectionFieldDescription

	// Indexes contains the secondary indexes that this Collection has.
	Indexes []IndexDescription

	// Policy contains the policy information on this collection.
	//
	// It is possible for a collection to not have a policy, a collection
	// without a policy has no access control.
	//
	// Note: The policy information must be validated using acp right after
	// parsing is done, to avoid storing an invalid policyID or policy resource
	// that may not even exist on acp.
	Policy immutable.Option[PolicyDescription]

	// IsMaterialized defines whether the items in this collection are cached or not.
	//
	// If it is true, they will be, if false, the data returned on query will be calculated
	// at query-time from source.
	//
	// At the moment this can only be set to `false` if this collection sources its data from
	// another collection/query (is a View).
	IsMaterialized bool
}

// QuerySource represents a collection data source from a query.
//
// The query will be executed when data from this source is requested, and the query results
// yielded to the consumer.
type QuerySource struct {
	// Query contains the base query of this data source.
	Query request.Select

	// Transform is a optional Lens configuration.  If specified, data drawn from the [Query] will have the
	// transform applied before being returned.
	//
	// The transform is not limited to just transforming the input documents, it may also yield new ones, or filter out
	// those passed in from the underlying query.
	Transform immutable.Option[model.Lens]
}

// CollectionSource represents a collection data source from another collection instance.
//
// Data against all collection instances in a CollectionSource chain will be returned as-if
// from the same dataset when queried.  Lens transforms may be applied between instances.
//
// Typically these are used to link together multiple schema versions into the same dataset.
type CollectionSource struct {
	// SourceCollectionID is the local identifier of the source [CollectionDescription] from which to
	// share data.
	//
	// This is a bi-directional relationship, and documents in the host collection instance will also
	// be available to the source collection instance.
	SourceCollectionID uint32

	// Transform is a optional Lens configuration.  If specified, data drawn from the source will have the
	// transform applied before being returned by any operation on the host collection instance.
	//
	// If the transform supports an inverse operation, that inverse will be applied when the source collection
	// draws data from this host.
	Transform immutable.Option[model.Lens]
}

// IDString returns the collection ID as a string.
func (col CollectionDescription) IDString() string {
	return fmt.Sprint(col.ID)
}

// GetFieldByName returns the field for the given field name. If such a field is found it
// will return it and true, if it is not found it will return false.
func (col CollectionDescription) GetFieldByName(fieldName string) (CollectionFieldDescription, bool) {
	for _, field := range col.Fields {
		if field.Name == fieldName {
			return field, true
		}
	}
	return CollectionFieldDescription{}, false
}

// GetFieldByRelation returns the field that supports the relation of the given name.
func (col CollectionDescription) GetFieldByRelation(
	relationName string,
	otherCollectionName string,
	otherFieldName string,
) (CollectionFieldDescription, bool) {
	for _, field := range col.Fields {
		if field.RelationName.Value() == relationName &&
			!(col.Name.Value() == otherCollectionName && otherFieldName == field.Name) &&
			field.Kind.Value() != FieldKind_DocID {
			return field, true
		}
	}
	return CollectionFieldDescription{}, false
}

// QuerySources returns all the Sources of type [QuerySource]
func (col CollectionDescription) QuerySources() []*QuerySource {
	return sourcesOfType[*QuerySource](col)
}

// CollectionSources returns all the Sources of type [CollectionSource]
func (col CollectionDescription) CollectionSources() []*CollectionSource {
	return sourcesOfType[*CollectionSource](col)
}

func sourcesOfType[ResultType any](col CollectionDescription) []ResultType {
	result := []ResultType{}
	for _, source := range col.Sources {
		if typedSource, isOfType := source.(ResultType); isOfType {
			result = append(result, typedSource)
		}
	}
	return result
}

// collectionDescription is a private type used to facilitate the unmarshalling
// of json to a [CollectionDescription].
type collectionDescription struct {
	// These properties are unmarshalled using the default json unmarshaller
	Name            immutable.Option[string]
	ID              uint32
	RootID          uint32
	SchemaVersionID string
	IsMaterialized  bool
	Policy          immutable.Option[PolicyDescription]
	Indexes         []IndexDescription
	Fields          []CollectionFieldDescription

	// Properties below this line are unmarshalled using custom logic in [UnmarshalJSON]
	Sources []map[string]json.RawMessage
}

func (c *CollectionDescription) UnmarshalJSON(bytes []byte) error {
	var descMap collectionDescription
	err := json.Unmarshal(bytes, &descMap)
	if err != nil {
		return err
	}

	c.Name = descMap.Name
	c.ID = descMap.ID
	c.RootID = descMap.RootID
	c.SchemaVersionID = descMap.SchemaVersionID
	c.IsMaterialized = descMap.IsMaterialized
	c.Indexes = descMap.Indexes
	c.Fields = descMap.Fields
	c.Sources = make([]any, len(descMap.Sources))
	c.Policy = descMap.Policy

	for i, source := range descMap.Sources {
		sourceJson, err := json.Marshal(source)
		if err != nil {
			return err
		}

		var sourceValue any
		// We detect which concrete type each `Source` object is by detecting
		// non-nillable fields, if the key is present it must be of that type.
		// They must be non-nillable as nil values may have their keys omitted from
		// the json. This also relies on the fields being unique.  We may wish to change
		// this later to custom-serialize with a `_type` property.
		if _, ok := source["Query"]; ok {
			// This must be a QuerySource, as only the `QuerySource` type has a `Query` field
			var querySource QuerySource
			err := json.Unmarshal(sourceJson, &querySource)
			if err != nil {
				return err
			}
			sourceValue = &querySource
		} else if _, ok := source["SourceCollectionID"]; ok {
			// This must be a CollectionSource, as only the `CollectionSource` type has a `SourceCollectionID` field
			var collectionSource CollectionSource
			err := json.Unmarshal(sourceJson, &collectionSource)
			if err != nil {
				return err
			}
			sourceValue = &collectionSource
		} else {
			return ErrFailedToUnmarshalCollection
		}

		c.Sources[i] = sourceValue
	}

	return nil
}
