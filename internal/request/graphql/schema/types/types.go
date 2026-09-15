// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// Package types contains the stable GraphQL names shared by schema intake and
// request parsing. Executable GraphQL types are defined in the canonical SDL
// templates rather than constructed in Go.
package types

import (
	"regexp"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
)

const (
	ExplainEnumDescription     = "ExplainType is an enum selecting the type of explanation done by the @explain directive."
	SimilarityFieldDescription = `
Returns how similar the given vector is to the specified field's value. The metric is
 whichever one the field's vector index was created with, defaulting to cosine similarity
 when the field has no vector index. Higher values are more similar for every metric.
`

	ExplainLabel    = "explain"
	ExhaustiveLabel = "exhaustive"
	PrimaryLabel    = "primary"
	RelationLabel   = "relation"

	ExplainArgNameType = "type"
	ExplainArgSimple   = "simple"
	ExplainArgExecute  = "execute"
	ExplainArgDebug    = "debug"

	CRDTDirectiveLabel    = "crdt"
	CRDTDirectivePropType = "type"

	ConstraintsDirectiveLabel    = "constraints"
	ConstraintsDirectivePropSize = "size"

	VectorEmbeddingDirectiveLabel        = "embedding"
	VectorEmbeddingDirectivePropProvider = "provider"
	VectorEmbeddingDirectivePropModel    = "model"
	VectorEmbeddingDirectivePropURL      = "url"
	VectorEmbeddingDirectivePropFields   = "fields"
	VectorEmbeddingDirectivePropTemplate = "template"

	PolicySchemaDirectiveLabel        = "policy"
	PolicySchemaDirectivePropID       = "id"
	PolicySchemaDirectivePropResource = "resource"

	IndexDirectiveLabel         = "index"
	IndexDirectivePropName      = "name"
	IndexDirectivePropUnique    = "unique"
	IndexDirectivePropDirection = "direction"
	IndexDirectivePropIncludes  = "includes"
	IndexDirectivePropKind      = "kind"

	OrderedIndexKind = "ordered"

	EncryptedIndexDirectiveLabel    = "encryptedIndex"
	EncryptedIndexDirectivePropType = "type"

	VectorIndexKind             = "vector"
	VectorIndexPropDimensions   = "dimensions"
	VectorIndexPropAlgorithm    = "alg"
	VectorIndexPropHNSW         = "hnsw"
	VectorIndexAlgorithmHNSW    = "hnsw"
	VectorIndexConfigPropMetric = "metric"

	VectorIndexHNSWConfigPropM              = "M"
	VectorIndexHNSWConfigPropEfConstruction = "efConstruction"
	VectorIndexHNSWConfigPropEfSearch       = "efSearch"

	VectorDistanceMetricCosine    = string(client.DistanceMetricCosine)
	VectorDistanceMetricEuclidean = string(client.DistanceMetricEuclidean)
	VectorDistanceMetricDot       = string(client.DistanceMetricDotProduct)

	IncludesPropField     = "field"
	IncludesPropDirection = "direction"

	DefaultDirectiveLabel     = "default"
	DefaultDirectivePropValue = "value"

	MaterializedDirectiveLabel  = "materialized"
	MaterializedDirectivePropIf = "if"

	BranchableDirectiveLabel  = "branchable"
	BranchableDirectivePropIf = "if"

	FieldOrderASC  = "ASC"
	FieldOrderDESC = "DESC"

	SimilarityArgVector = "vector"
)

type enumDescription struct{ description string }

func (e *enumDescription) Description() string { return e.description }

// ExplainEnum retains the small description API used by clients without
// reintroducing the legacy executable GraphQL type dependency.
func ExplainEnum() *enumDescription {
	return &enumDescription{description: ExplainEnumDescription}
}

var BlobPattern = regexp.MustCompile("^[0-9a-fA-F]+$")

// ParseCRDTType resolves the public GraphQL enum spelling to its DefraDB CType.
func ParseCRDTType(name string) (client.CType, bool) {
	for _, fieldCRDT := range crdt.FieldCRDTs {
		if fieldCRDT.String() == name {
			return fieldCRDT.CType(), true
		}
	}
	return 0, false
}
