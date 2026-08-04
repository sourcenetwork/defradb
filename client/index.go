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
)

// IndexFieldDescription describes how a field is being indexed.
type IndexedFieldDescription struct {
	// Name contains the name of the field.
	Name string
	// Descending indicates whether the field is indexed in descending order.
	Descending bool
}

// IndexKind identifies the kind of an index. It is the single authority on what kind an
// [IndexDescription] is: it is a stored field, not derived from which sub-descriptor is set, so
// there is no way for two sub-descriptors to disagree about the kind. The zero value is
// [IndexKindOrdered], so a descriptor persisted before this field existed reads back as an
// ordered index.
type IndexKind uint8

const (
	// IndexKindOrdered identifies an ordered index: one that stores field values as order-preserving
	// keys (the scalar/unique/JSON indexes). This is the zero value so that legacy descriptors, which
	// predate the kind distinction, default to it.
	IndexKindOrdered IndexKind = iota
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

	// Upper bounds, checked at index creation. Build cost grows with these, and any client can set
	// them when Node-ACP is off (the default), so an unbounded value makes one write do unbounded
	// work. Set well above any useful value.
	MaxHNSWM              uint32 = 512
	MaxHNSWEfConstruction uint32 = 4096
	MaxHNSWEfSearch       uint32 = 4096
)

// VectorIndexDescription holds config specific to vector (ANN) indexes.
type VectorIndexDescription struct {
	// Algorithm is the algorithm used to build/search the index.
	Algorithm VectorAlgorithm
	// Metric is the distance metric used to compare vectors.
	Metric DistanceMetric
	// Dimensions is the length of the vectors being indexed. It must be set, except on an @embedding
	// field, where the embedding model fixes the length and Dimensions may be left 0.
	Dimensions uint32
	// HNSW holds HNSW-specific parameters. Non-nil when Algorithm == VectorAlgorithmHNSW.
	HNSW *HNSWParams
}

// IndexDescription describes an index.
//
// Kind is the authority on the index kind. The kind-specific sub-descriptor (e.g. [Vector]) is set
// alongside it, but Kind is what callers switch on, so adding a new kind cannot introduce an
// ambiguous "two sub-descriptors set" state.
type IndexDescription struct {
	// Name contains the name of the index.
	Name string
	// ID is the local identifier of this index.
	ID uint32
	// Fields contains the fields that are being indexed.
	Fields []IndexedFieldDescription
	// Kind is the kind of this index. It is the single authority on the kind; the zero value is
	// [IndexKindOrdered], so descriptors persisted before this field existed read back as ordered.
	Kind IndexKind
	// Unique indicates whether the index is unique. It only applies to ordered indexes.
	Unique bool
	// Vector holds config specific to vector (ANN) indexes. Non-nil when Kind is [IndexKindVector].
	Vector *VectorIndexDescription
}

// IsVector returns true if this is a vector index.
func (d IndexDescription) IsVector() bool {
	return d.Kind == IndexKindVector
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
