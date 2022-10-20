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

type (
	OrderDirection string

	SelectionType int

	OrderCondition struct {
		// field may be a compound field statement
		// since the order statement allows ordering on
		// sub objects.
		//
		// Given the statement: {order: {author: {birthday: DESC}}}
		// The field value would be "author.birthday"
		// and the direction would be "DESC"
		Fields    []string
		Direction OrderDirection
	}

	GroupBy struct {
		Fields []string
	}

	OrderBy struct {
		Conditions []OrderCondition
	}
)

const (
	// GQL special field, returns the host object's type name
	// https://spec.graphql.org/October2021/#sec-Type-Name-Introspection
	TypeNameFieldName = "__typename"

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
	DepthClause   = "depth"

	AverageFieldName = "_avg"
	CountFieldName   = "_count"
	DocKeyFieldName  = "_key"
	GroupFieldName   = "_group"
	SumFieldName     = "_sum"
	VersionFieldName = "_version"

	ExplainLabel = "explain"

	LatestCommitsQueryName = "latestCommits"
	CommitsQueryName       = "commits"

	CommitTypeName  = "Commit"
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
		TypeNameFieldName: true,
		VersionFieldName:  true,
		GroupFieldName:    true,
		CountFieldName:    true,
		SumFieldName:      true,
		AverageFieldName:  true,
		DocKeyFieldName:   true,
	}

	Aggregates = map[string]struct{}{
		CountFieldName:   {},
		SumFieldName:     {},
		AverageFieldName: {},
	}

	CommitQueries = map[string]struct{}{
		LatestCommitsQueryName: {},
		CommitsQueryName:       {},
	}

	VersionFields = []string{
		HeightFieldName,
		CidFieldName,
		DeltaFieldName,
	}

	LinksFields = []string{
		LinksNameFieldName,
		LinksCidFieldName,
	}
)
