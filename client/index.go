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

// IndexKind is the config for one kind of index. The value is either an [*OrderedIndexDescription]
// or a [*VectorIndexDescription], and the concrete type IS the kind, so an index can never be in an
// invalid "kind says one thing, config says another" state. This mirrors [FieldKind].
type IndexKind interface {
	// String returns the name of the kind.
	String() string

	// isIndexKind is unexported so only types in this package can be an IndexKind.
	isIndexKind()
}

// VectorAlgorithm identifies the algorithm used to build/search a vector index. It is a string so it
// serializes as a readable name (e.g. in a stored descriptor or the CLI's --vector JSON) rather than
// an opaque number.
type VectorAlgorithm string

const (
	// VectorAlgorithmHNSW identifies the HNSW (Hierarchical Navigable Small World) algorithm.
	VectorAlgorithmHNSW VectorAlgorithm = "HNSW"
)

// DistanceMetric identifies the distance metric used to compare vectors. It is a string so it
// serializes as a readable name (e.g. in a stored descriptor or the CLI's --vector JSON) rather than
// an opaque number.
type DistanceMetric string

const (
	// DistanceMetricCosine identifies the cosine distance metric.
	DistanceMetricCosine DistanceMetric = "COSINE"
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

// OrderedIndexDescription is the config for an ordered index: one that stores field values as
// order-preserving keys (the scalar/unique/JSON indexes).
type OrderedIndexDescription struct {
	// Unique indicates whether the index enforces uniqueness.
	Unique bool
}

func (OrderedIndexDescription) String() string { return "Ordered" }
func (*OrderedIndexDescription) isIndexKind()  {}

var _ IndexKind = (*OrderedIndexDescription)(nil)

// VectorIndexDescription is the config for a vector (ANN) index.
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

func (VectorIndexDescription) String() string { return "Vector" }
func (*VectorIndexDescription) isIndexKind()  {}

var _ IndexKind = (*VectorIndexDescription)(nil)

// IndexDescription describes an index.
type IndexDescription struct {
	// Name contains the name of the index.
	Name string
	// ID is the local identifier of this index.
	ID uint32
	// Fields contains the fields that are being indexed.
	Fields []IndexedFieldDescription

	// Unique indicates whether the index enforces uniqueness. It only applies to ordered indexes.
	//
	// It is kept for backward compatibility with callers of the embedded API that predate [Kind]:
	// they still set and read this field directly. It mirrors the ordered [Kind]'s uniqueness and is
	// kept in sync at the (un)marshalling and construction boundaries. Prefer [GetUnique] in new code.
	Unique bool

	// Kind is the kind-specific config: either an [*OrderedIndexDescription] or a
	// [*VectorIndexDescription]. The concrete type is the kind, so it cannot disagree with itself, and
	// it is the authority the rest of the system dispatches on (via [IsVector]/[GetVector]/[GetUnique]).
	Kind IndexKind
}

// IsVector returns true if this is a vector index.
func (d IndexDescription) IsVector() bool {
	_, ok := d.Kind.(*VectorIndexDescription)
	return ok
}

// GetVector returns the vector config and true if this is a vector index, otherwise nil and false.
func (d IndexDescription) GetVector() (*VectorIndexDescription, bool) {
	v, ok := d.Kind.(*VectorIndexDescription)
	return v, ok
}

// GetUnique returns whether the index is unique. Only ordered indexes can be unique.
func (d IndexDescription) GetUnique() bool {
	if o, ok := d.Kind.(*OrderedIndexDescription); ok {
		return o.Unique
	}
	// Kind not yet resolved (a caller built the struct with only the legacy Unique field): fall back
	// to it. A vector index never sets Unique, so this stays correct.
	return d.Unique
}

// Normalize returns a copy with Kind and the compat Unique field consistent: a nil Kind is filled
// from Unique (an ordered index), and Unique is set from the resolved Kind. Use it to compare two
// descriptors that may have been built in different styles (only Unique, or only Kind).
func (d IndexDescription) Normalize() IndexDescription {
	if d.Kind == nil {
		d.Kind = &OrderedIndexDescription{Unique: d.Unique}
	}
	d.Unique = d.GetUnique()
	return d
}

// indexDescription mirrors [IndexDescription] for unmarshalling: Kind is read as raw JSON and
// resolved to a concrete type by [parseIndexKind].
type indexDescription struct {
	Name   string
	ID     uint32
	Fields []IndexedFieldDescription
	// Kind is unmarshalled with custom logic in [UnmarshalJSON].
	Kind json.RawMessage
	// Unique is read only for backward compatibility: descriptors written before the interface
	// existed stored Unique at the top level. See [parseIndexKind].
	Unique bool
}

func (d *IndexDescription) UnmarshalJSON(bytes []byte) error {
	var mirror indexDescription
	if err := json.Unmarshal(bytes, &mirror); err != nil {
		return err
	}
	d.Name = mirror.Name
	d.ID = mirror.ID
	d.Fields = mirror.Fields
	kind, err := parseIndexKind(mirror.Kind, mirror.Unique)
	if err != nil {
		return err
	}
	d.Kind = kind
	// Keep the compat Unique field in sync with the resolved kind so old readers see it.
	d.Unique = d.GetUnique()
	return nil
}

// MarshalJSON writes both the compat top-level Unique (so a reader that predates [Kind] still gets
// the uniqueness) and the Kind object (for readers that use it). They are always emitted consistently.
func (d IndexDescription) MarshalJSON() ([]byte, error) {
	kind := d.Kind
	if kind == nil {
		// The struct was built with only the legacy Unique field; derive the kind from it.
		kind = &OrderedIndexDescription{Unique: d.Unique}
	}
	unique := false
	if o, ok := kind.(*OrderedIndexDescription); ok {
		unique = o.Unique
	}
	return json.Marshal(indexDescription{
		Name:   d.Name,
		ID:     d.ID,
		Fields: d.Fields,
		Kind:   mustMarshalKind(kind),
		Unique: unique,
	})
}

func mustMarshalKind(kind IndexKind) json.RawMessage {
	// A concrete kind is a plain struct; its Marshal cannot fail. Ignore the error to keep the
	// signature clean, the same way the concrete kinds are trivially serialisable.
	b, _ := json.Marshal(kind)
	return b
}

// parseIndexKind resolves the raw Kind JSON to a concrete [IndexKind]. A missing Kind is a legacy
// descriptor (written before the interface): it is an ordered index, and any top-level Unique it had
// is carried into the ordered config. A present Kind is recognised by its shape — a vector config has
// an Algorithm/Metric/Dimensions, an ordered config does not.
func parseIndexKind(raw json.RawMessage, legacyUnique bool) (IndexKind, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return &OrderedIndexDescription{Unique: legacyUnique}, nil
	}
	// Sniff for a vector config: it has a Dimensions field, which the ordered config does not.
	var probe struct {
		Algorithm  *VectorAlgorithm
		Dimensions *uint32
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, err
	}
	if probe.Algorithm != nil || probe.Dimensions != nil {
		var v VectorIndexDescription
		if err := json.Unmarshal(raw, &v); err != nil {
			return nil, err
		}
		return &v, nil
	}
	var o OrderedIndexDescription
	if err := json.Unmarshal(raw, &o); err != nil {
		return nil, err
	}
	return &o, nil
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
