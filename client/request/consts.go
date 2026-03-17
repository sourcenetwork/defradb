// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package request

const (
	// GQL special field, returns the host object's type name
	// https://spec.graphql.org/October2021/#sec-Type-Name-Introspection
	TypeNameFieldName = "__typename"

	Input              = "input"
	AddInput           = "add"
	UpdateInput        = "update"
	FieldArgName       = "field"
	FieldIDName        = "fieldId"
	FieldNameName      = "fieldName"
	CompositeFieldName = "_C"
	ShowDeleted        = "showDeleted"

	EncryptDocArgName         = "encrypt"
	EncryptFieldsArgName      = "encryptFields"
	EncryptedCollectionPrefix = "encrypted_"
	EncryptedSearchResultName = "EncryptedSearchResult"

	FilterClause  = "filter"
	GroupByClause = "groupBy"
	LimitClause   = "limit"
	OffsetClause  = "offset"
	OrderClause   = "order"
	DepthClause   = "depth"

	DocIDArgName  = "docID"
	CidArgName    = "cid"
	HeightArgName = "height"

	DocIDFieldName   = "_docID"
	DeletedFieldName = "_deleted"
	VersionFieldName = "_version"
	AliasFieldName   = "_alias"

	MaxFieldName        = "MAX"
	MinFieldName        = "MIN"
	SimilarityFieldName = "SIMILARITY"
	SumFieldName        = "SUM"
	GroupFieldName      = "GROUP"
	AverageFieldName    = "AVG"
	CountFieldName      = "COUNT"

	// New generated document id from a backed up document,
	// which might have a different _docID originally.
	NewDocIDFieldName = "_docIDNew"

	ExplainLabel    = "explain"
	ExhaustiveLabel = "exhaustive"

	CommitsName = "_commits"

	CommitTypeName               = "Commit"
	LinksFieldName               = "links"
	HeadsFieldName               = "heads"
	SignatureFieldName           = "signature"
	SignatureTypeName            = "Signature"
	HeightFieldName              = "height"
	CollectionVersionIDFieldName = "collectionVersionId"
	DeltaFieldName               = "delta"

	// SelfTypeName is the name given to relation field types that reference the host type.
	//
	// For example, when a `User` collection contains a relation to the `User` collection the field
	// will be of type [SelfTypeName].
	SelfTypeName = "Self"

	LinksNameFieldName = "linkName"
	CidFieldName       = "cid"

	SignatureTypeFieldName     = "type"
	SignatureIdentityFieldName = "identity"
	SignatureValueFieldName    = "value"

	DocIDsFieldName = "docIDs"

	ASC  = OrderDirection("ASC")
	DESC = OrderDirection("DESC")
)

var AggregateFields = []string{
	MaxFieldName,
	MinFieldName,
	SimilarityFieldName,
	SumFieldName,
	GroupFieldName,
	AverageFieldName,
	CountFieldName,
}

var (
	NameToOrderDirection = map[string]OrderDirection{
		string(ASC):  ASC,
		string(DESC): DESC,
	}

	// ReservedTypeNames is the set of type names reserved by the system.
	//
	// Users cannot define types using these names.
	//
	// For example, collections may not be defined using these names.
	ReservedTypeNames = map[string]struct{}{
		SelfTypeName: {},
	}

	ReservedFields = map[string]struct{}{
		TypeNameFieldName:   {},
		VersionFieldName:    {},
		GroupFieldName:      {},
		CountFieldName:      {},
		SumFieldName:        {},
		AverageFieldName:    {},
		DocIDFieldName:      {},
		DeletedFieldName:    {},
		MaxFieldName:        {},
		MinFieldName:        {},
		SimilarityFieldName: {},
	}

	Aggregates = map[string]struct{}{
		CountFieldName:   {},
		SumFieldName:     {},
		AverageFieldName: {},
		MaxFieldName:     {},
		MinFieldName:     {},
	}

	VersionFields = []string{
		// DocIDArgName must be the first in this slice in order to align with document doc-id mappings
		DocIDArgName,
		HeightFieldName,
		CidFieldName,
		CollectionVersionIDFieldName,
		FieldNameName,
		DeltaFieldName,
		LinksNameFieldName,
	}

	LinksFields = []string{
		LinksNameFieldName,
		CidFieldName,
	}

	SignatureFields = []string{
		SignatureTypeFieldName,
		SignatureIdentityFieldName,
		SignatureValueFieldName,
	}
)

// This is appended to the related object name to give us the field name
// that corresponds to the related object's join relation id, i.e. `_authorID`.
const relatedObjectIDSuffix = "ID"

// ToFieldID converts a field name to its relation ID field name.
// For example: "author" becomes "_authorID"
func ToFieldID(fieldName string) string {
	return "_" + fieldName + relatedObjectIDSuffix
}

// ToRelatedObjectName extracts the object field name from a relation ID field name.
// For example: "_authorID" returns ("author", true)
// Returns ("", false) if the field name is not a valid relation ID field.
func ToRelatedObjectName(fieldName string) (string, bool) {
	if len(fieldName) <= len(relatedObjectIDSuffix)+1 {
		return "", false
	}
	if fieldName[0] != '_' {
		return "", false
	}
	if fieldName == DocIDFieldName {
		return "", false
	}
	if fieldName[len(fieldName)-len(relatedObjectIDSuffix):] != relatedObjectIDSuffix {
		return "", false
	}
	return fieldName[1 : len(fieldName)-len(relatedObjectIDSuffix)], true
}
