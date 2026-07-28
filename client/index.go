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
	"encoding/json"
)

// IndexFieldDescription describes how a field is being indexed.
type IndexedFieldDescription struct {
	// Name contains the name of the field.
	Name string
	// Descending indicates whether the field is indexed in descending order.
	Descending bool
}

// IndexKind identifies the kind of an index.
type IndexKind uint8

const (
	// IndexKindSecondary identifies a secondary (scalar/unique/JSON) index.
	IndexKindSecondary IndexKind = iota
	// IndexKindVector identifies a vector (ANN) index.
	IndexKindVector
)

// VectorAlgorithm identifies the algorithm used to build/search a vector index.
type VectorAlgorithm uint8

const (
	// VectorAlgorithmHNSW identifies the HNSW (Hierarchical Navigable Small World) algorithm.
	VectorAlgorithmHNSW VectorAlgorithm = iota
)

// DistanceMetric identifies the distance metric used to compare vectors.
type DistanceMetric uint8

const (
	// DistanceMetricCosine identifies the cosine distance metric.
	DistanceMetricCosine DistanceMetric = iota
)

// SecondaryIndexDescription holds config specific to secondary (scalar/unique/JSON) indexes.
type SecondaryIndexDescription struct {
	// Unique indicates whether the index enforces uniqueness.
	Unique bool
}

// HNSWParams holds HNSW-specific build/search parameters.
type HNSWParams struct {
	// M is the maximum number of connections per node.
	M uint32
	// EfConstruction controls the size of the dynamic candidate list during index construction.
	EfConstruction uint32
	// EfSearch controls the size of the dynamic candidate list during search.
	EfSearch uint32
}

// Default HNSW parameters, applied when the corresponding @vectorIndex directive argument is
// omitted. These are the single source of truth: both the GraphQL directive definition and the
// directive parser reference them, so the documented defaults cannot drift apart.
const (
	// DefaultHNSWM is the default maximum number of connections per node. Higher values improve
	// recall at the cost of memory and build time.
	DefaultHNSWM uint32 = 16
	// DefaultHNSWEfConstruction is the default build-time exploration factor. Higher values improve
	// graph quality (recall) at the cost of build time.
	DefaultHNSWEfConstruction uint32 = 128
	// DefaultHNSWEfSearch is the default query-time exploration factor. Higher values improve recall
	// at the cost of query latency; it may be overridden per query.
	DefaultHNSWEfSearch uint32 = 64
)

// VectorIndexDescription holds config specific to vector (ANN) indexes.
type VectorIndexDescription struct {
	// Algorithm is the algorithm used to build/search the index.
	Algorithm VectorAlgorithm
	// Metric is the distance metric used to compare vectors.
	Metric DistanceMetric
	// Dimensions is the number of dimensions of the vectors being indexed.
	Dimensions uint32
	// HNSW holds HNSW-specific parameters. Non-nil when Algorithm == VectorAlgorithmHNSW.
	HNSW *HNSWParams
}

// IndexDescription describes an index.
//
// Exactly one of [Secondary] or [Vector] is non-nil; the non-nil one determines the index
// kind (see [IndexDescription.Kind]). A descriptor with neither set is treated as a secondary
// index for backward compatibility with descriptors persisted before this distinction existed.
type IndexDescription struct {
	// Name contains the name of the index.
	Name string
	// ID is the local identifier of this index.
	ID uint32
	// Fields contains the fields that are being indexed.
	Fields []IndexedFieldDescription
	// Secondary holds config specific to secondary (scalar/unique/JSON) indexes. Non-nil iff
	// this is a secondary index.
	Secondary *SecondaryIndexDescription
	// Vector holds config specific to vector (ANN) indexes. Non-nil iff this is a vector index.
	Vector *VectorIndexDescription
}

// Kind returns the index kind, derived from which sub-struct is set. A descriptor with
// neither [Secondary] nor [Vector] set is treated as [IndexKindSecondary] (legacy descriptors).
func (d IndexDescription) Kind() IndexKind {
	if d.Vector != nil {
		return IndexKindVector
	}
	return IndexKindSecondary
}

// indexDescription is a private type used to facilitate the marshalling and unmarshalling
// of json to/from an [IndexDescription].
//
// Existing persisted descriptors store `Unique` at the top level (no nested `Secondary`
// struct), so [UnmarshalJSON] detects and migrates that legacy shape.
type indexDescription struct {
	Name      string
	ID        uint32
	Fields    []IndexedFieldDescription
	Secondary *SecondaryIndexDescription
	Vector    *VectorIndexDescription

	// Unique is the legacy top-level location of the secondary index's uniqueness flag. It is
	// only ever read here; new descriptors are marshalled with `Unique` nested under `Secondary`.
	Unique *bool
}

func (d *IndexDescription) UnmarshalJSON(bytes []byte) error {
	var descMap indexDescription
	err := json.Unmarshal(bytes, &descMap)
	if err != nil {
		return err
	}

	d.Name = descMap.Name
	d.ID = descMap.ID
	d.Fields = descMap.Fields
	d.Vector = descMap.Vector

	switch {
	case descMap.Secondary != nil:
		d.Secondary = descMap.Secondary
	case descMap.Unique != nil:
		d.Secondary = &SecondaryIndexDescription{Unique: *descMap.Unique}
	case descMap.Vector == nil:
		d.Secondary = &SecondaryIndexDescription{}
	}

	return nil
}

func (d IndexDescription) MarshalJSON() ([]byte, error) {
	return json.Marshal(indexDescription{
		Name:      d.Name,
		ID:        d.ID,
		Fields:    d.Fields,
		Secondary: d.Secondary,
		Vector:    d.Vector,
	})
}

// NewIndexRequest describes an index creation request.
// It does not contain the ID, as it is not a valid field for the request body.
// Instead it should be automatically generated.
type NewIndexRequest struct {
	// Name contains the name of the index.
	Name string
	// Fields contains the fields that are being indexed.
	Fields []IndexedFieldDescription
	// Unique indicates whether the index is unique.
	Unique bool
	// Vector holds config specific to vector (ANN) indexes. Non-nil iff this is a vector index request.
	Vector *VectorIndexDescription
}

// CollectionIndex is an interface for indexing documents in a collection.
type CollectionIndex interface {
	// Save indexes a document by storing indexed field values.
	// It doesn't retire previous values. For this [Update] should be used.
	Save(context.Context, *Document) error
	// Update updates an existing document in the index.
	// It removes the previous indexed field values and stores the new ones.
	Update(context.Context, *Document, *Document) error
	// Delete deletes an existing document from the index
	Delete(context.Context, *Document) error
	// Name returns the name of the index
	Name() string
	// Description returns the description of the index
	Description() IndexDescription
}

// ListIndexesResult is a single entry returned by ListIndexes: an index's static description
// together with its runtime lifecycle, kept as separate values rather than blended into one.
//
// Execution reports the lifecycle through the action system (Execution.Action and
// Execution.Status):
//   - building: InProgress + BackfillIndexAction
//   - failed:   Errored    + BackfillIndexAction (Execution.Reason has the cause)
//   - ready:    Completed  (no in-flight action; Execution.Action is unset)
//
// A whole-index drop is not reported here: the index is removed from the listing when the drop is
// requested, so a dropping index is simply absent rather than shown with a status.
type ListIndexesResult struct {
	// CollectionName is the name of the collection the index belongs to.
	CollectionName string
	// Description is the static index specification.
	Description IndexDescription
	// Execution is the index's current (or most recent) lifecycle action.
	Execution ActionExecution
}

// CollectIndexedFields returns all fields that are indexed by all collection indexes.
func (col CollectionVersion) CollectIndexedFields() []CollectionFieldDescription {
	fieldsMap := make(map[string]bool)
	fields := make([]CollectionFieldDescription, 0, len(col.Indexes))
	for _, index := range col.Indexes {
		for _, field := range index.Fields {
			if fieldsMap[field.Name] {
				// If the FieldDescription has already been added to the result do not add it a second time
				// this can happen if a field is referenced by multiple indexes
				continue
			}
			colField, ok := col.GetFieldByName(field.Name)
			if ok {
				fields = append(fields, colField)
			}
		}
	}
	return fields
}

// GetIndexesOnField returns all indexes that are indexing the given field.
// If the field is not the first field of a composite index, the index is not returned.
func (col CollectionVersion) GetIndexesOnField(fieldName string) []IndexDescription {
	result := []IndexDescription{}
	for _, index := range col.Indexes {
		if index.Fields[0].Name == fieldName {
			result = append(result, index)
		}
	}
	return result
}
