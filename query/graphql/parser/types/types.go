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
	Cid     = string("cid")
	Data    = string("data")
	DocKey  = string("dockey")
	DocKeys = string("dockeys")
	Field   = string("field")
	Id      = string("id")
	Ids     = string("ids")

	FilterClause  = string("filter")
	GroupByClause = string("groupBy")
	LimitClause   = string("limit")
	OffsetClause  = string("offset")
	OrderClause   = string("order")

	ASC  = SortDirection("ASC")
	DESC = SortDirection("DESC")

	VersionFieldName = "_version"
	GroupFieldName   = "_group"
	DocKeyFieldName  = "_key"
	CountFieldName   = "_count"
	SumFieldName     = "_sum"
	AverageFieldName = "_avg"
	HiddenFieldName  = "_hidden"

	ScanQuery = iota
	VersionedScanQuery

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
		HiddenFieldName:  true,
		DocKeyFieldName:  true,
	}

	Aggregates = map[string]struct{}{
		CountFieldName:   {},
		SumFieldName:     {},
		AverageFieldName: {},
	}
)
