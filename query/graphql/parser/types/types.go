// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package types defines the GraphQL types used by the query service.
*/
package types

import "github.com/graphql-go/graphql/language/ast"

type (
	OrderDirection string

	SelectionType int

	// Enum for different types of read Select queries
	SelectQueryType int

	OrderCondition struct {
		// field may be a compound field statement
		// since the order statement allows ordering on
		// sub objects.
		//
		// Given the statement: {order: {author: {birthday: DESC}}}
		// The field value would be "author.birthday"
		// and the direction would be "DESC"
		Field     string
		Direction OrderDirection
	}

	GroupBy struct {
		Fields []string
	}

	OrderBy struct {
		Conditions []OrderCondition
		Statement  *ast.ObjectValue
	}

	Limit struct {
		Limit  int64
		Offset int64
	}

	OptionalDocKeys struct {
		HasValue bool
		Value    []string
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

	ASC  = OrderDirection("ASC")
	DESC = OrderDirection("DESC")
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
	NameToOrderDirection = map[string]OrderDirection{
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
