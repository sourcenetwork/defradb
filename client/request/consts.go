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

	// This is appended to the related object name to give us the field name
	// that corresponds to the related object's join relation id, i.e. `Author_id`.
	RelatedObjectID = "_id"

	Input              = "input"
	CreateInput        = "create"
	UpdateInput        = "update"
	FieldArgName       = "field"
	FieldNameArgName   = "fieldName"
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

	AverageFieldName    = "_avg"
	CountFieldName      = "_count"
	DocIDFieldName      = "_docID"
	GroupFieldName      = "_group"
	DeletedFieldName    = "_deleted"
	SumFieldName        = "_sum"
	VersionFieldName    = "_version"
	MaxFieldName        = "_max"
	MinFieldName        = "_min"
	AliasFieldName      = "_alias"
	SimilarityFieldName = "_similarity"

	// New generated document id from a backed up document,
	// which might have a different _docID originally.
	NewDocIDFieldName = "_docIDNew"

	ExplainLabel = "explain"

	CommitsName = "_commits"

	CommitTypeName           = "Commit"
	LinksFieldName           = "links"
	SignatureFieldName       = "signature"
	SignatureTypeName        = "Signature"
	HeightFieldName          = "height"
	SchemaVersionIDFieldName = "schemaVersionId"
	DeltaFieldName           = "delta"

	DeltaArgFieldName       = "FieldName"
	DeltaArgData            = "Data"
	DeltaArgSchemaVersionID = "SchemaVersionID"
	DeltaArgPriority        = "Priority"
	DeltaArgDocID           = "DocID"

	// SelfTypeName is the name given to relation field types that reference the host type.
	//
	// For example, when a `User` collection contains a relation to the `User` collection the field
	// will be of type [SelfTypeName].
	SelfTypeName = "Self"

	LinksNameFieldName = "name"
	CidFieldName       = "cid"

	SignatureTypeFieldName     = "type"
	SignatureIdentityFieldName = "identity"
	SignatureValueFieldName    = "value"

	DocIDsFieldName = "docIDs"

	ASC  = OrderDirection("ASC")
	DESC = OrderDirection("DESC")
)

var (
	NameToOrderDirection = map[string]OrderDirection{
		string(ASC):  ASC,
		string(DESC): DESC,
	}

	// ReservedTypeNames is the set of type names reserved by the system.
	//
	// Users cannot define types using these names.
	//
	// For example, collections and schemas may not be defined using these names.
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
		HeightFieldName,
		CidFieldName,
		DocIDArgName,
		SchemaVersionIDFieldName,
		FieldNameName,
		DeltaFieldName,
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
