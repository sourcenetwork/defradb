// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package types

import "github.com/graphql-go/graphql/language/ast"

type (
	SortDirection string

	SelectionType int

	// Enum for different types of read Select queries
	SelectQueryType int

	SortCondition struct {
		// field may be a compound field statement
		// since the sort statement allows sorting on
		// sub objects.
		//
		// Given the statement: {sort: {author: {birthday: DESC}}}
		// The field value would be "author.birthday"
		// and the direction would be "DESC"
		Field     string
		Direction SortDirection
	}

	GroupBy struct {
		Fields []string
	}

	OrderBy struct {
		Conditions []SortCondition
		Statement  *ast.ObjectValue
	}

	Limit struct {
		Limit  int64
		Offset int64
	}
)

const (
	Cid     = "cid"
	Data    = "data"
	DocKey  = "dockey"
	DocKeys = "dockeys"
	Field   = "field"
	Id      = "id"
	Ids     = "ids"

	FilterClause  = "filter"
	GroupByClause = "groupBy"
	LimitClause   = "limit"
	OffsetClause  = "offset"
	OrderClause   = "order"

	AverageFieldName = "_avg"
	CountFieldName   = "_count"
	DocKeyFieldName  = "_key"
	GroupFieldName   = "_group"
	SumFieldName     = "_sum"
	VersionFieldName = "_version"

	ExplainLabel = "explain"

	LinksFieldName  = "links"
	HeightFieldName = "height"
	CidFieldName    = "cid"
	DeltaFieldName  = "delta"

	LinksNameFieldName = "name"
	LinksCidFieldName  = "cid"

	ASC  = SortDirection("ASC")
	DESC = SortDirection("DESC")
)

const (
	ScanQuery = iota
	VersionedScanQuery
)

const (
	NoneSelection = iota
	ObjectSelection
	CommitSelection
)

var (
	NameToSortDirection = map[string]SortDirection{
		string(ASC):  ASC,
		string(DESC): DESC,
	}

	ReservedFields = map[string]bool{
		VersionFieldName: true,
		GroupFieldName:   true,
		CountFieldName:   true,
		SumFieldName:     true,
		AverageFieldName: true,
		DocKeyFieldName:  true,
	}

	Aggregates = map[string]struct{}{
		CountFieldName:   {},
		SumFieldName:     {},
		AverageFieldName: {},
	}

	VersionFields = map[string]struct{}{
		HeightFieldName: {},
		CidFieldName:    {},
		DeltaFieldName:  {},
	}

	LinksFields = map[string]struct{}{
		LinksNameFieldName: {},
		LinksCidFieldName:  {},
	}
)
