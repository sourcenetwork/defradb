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

	Cid         = "cid"
	Input       = "input"
	Inputs      = "inputs"
	FieldName   = "field"
	FieldIDName = "fieldId"
	ShowDeleted = "showDeleted"

	EncryptArgName       = "encrypt"
	EncryptFieldsArgName = "encryptFields"

	FilterClause  = "filter"
	GroupByClause = "groupBy"
	LimitClause   = "limit"
	OffsetClause  = "offset"
	OrderClause   = "order"
	DepthClause   = "depth"

	DocIDArgName  = "docID"
	DocIDsArgName = "docIDs"

	AverageFieldName = "_avg"
	CountFieldName   = "_count"
	DocIDFieldName   = "_docID"
	GroupFieldName   = "_group"
	DeletedFieldName = "_deleted"
	SumFieldName     = "_sum"
	VersionFieldName = "_version"

	// New generated document id from a backed up document,
	// which might have a different _docID originally.
	NewDocIDFieldName = "_docIDNew"

	ExplainLabel = "explain"

	LatestCommitsName = "latestCommits"
	CommitsName       = "commits"

	CommitTypeName           = "Commit"
	LinksFieldName           = "links"
	HeightFieldName          = "height"
	CidFieldName             = "cid"
	CollectionIDFieldName    = "collectionID"
	SchemaVersionIDFieldName = "schemaVersionId"
	FieldNameFieldName       = "fieldName"
	FieldIDFieldName         = "fieldId"
	DeltaFieldName           = "delta"

	DeltaArgFieldName       = "FieldName"
	DeltaArgData            = "Data"
	DeltaArgSchemaVersionID = "SchemaVersionID"
	DeltaArgPriority        = "Priority"
	DeltaArgDocID           = "DocID"

	LinksNameFieldName = "name"
	LinksCidFieldName  = "cid"

	ASC  = OrderDirection("ASC")
	DESC = OrderDirection("DESC")
)

var (
	NameToOrderDirection = map[string]OrderDirection{
		string(ASC):  ASC,
		string(DESC): DESC,
	}

	ReservedFields = map[string]bool{
		TypeNameFieldName: true,
		VersionFieldName:  true,
		GroupFieldName:    true,
		CountFieldName:    true,
		SumFieldName:      true,
		AverageFieldName:  true,
		DocIDFieldName:    true,
		DeletedFieldName:  true,
	}

	Aggregates = map[string]struct{}{
		CountFieldName:   {},
		SumFieldName:     {},
		AverageFieldName: {},
	}

	CommitQueries = map[string]struct{}{
		LatestCommitsName: {},
		CommitsName:       {},
	}

	VersionFields = []string{
		HeightFieldName,
		CidFieldName,
		DocIDArgName,
		CollectionIDFieldName,
		SchemaVersionIDFieldName,
		FieldNameFieldName,
		FieldIDFieldName,
		DeltaFieldName,
	}

	LinksFields = []string{
		LinksNameFieldName,
		LinksCidFieldName,
	}
)
