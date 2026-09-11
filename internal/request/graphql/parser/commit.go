// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package parser

import (
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
)

func parseCommitSelect(
	exe *executionContext,
	field *field,
) (*request.CommitSelect, error) {
	commit := &request.CommitSelect{
		Field: request.Field{
			Name:  field.name,
			Alias: field.alias,
		},
	}

	for _, name := range field.argumentOrder {
		value := field.arguments[name]

		switch name {
		case request.DocIDArgName:
			var docIDs []string
			switch v := value.(type) {
			case []any:
				if len(v) > 1 {
					// todo - This limitiation is temporary and should be removed in
					// https://github.com/sourcenetwork/defradb/issues/4302
					return nil, ErrMultipleDocIDsNotSupported
				}

				docIDs = make([]string, len(v))
				for i, value := range v {
					docIDs[i] = value.(string)
				}

			case []string:
				if len(v) > 1 {
					// todo - This limitiation is temporary and should be removed in
					// https://github.com/sourcenetwork/defradb/issues/4302
					return nil, ErrMultipleDocIDsNotSupported
				}

				docIDs = v

			case string:
				docIDs = []string{v}

			default:
				continue
			}

			commit.DocIDs = immutable.Some(docIDs)

		case request.CidFieldName:
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}

			cids := make([]string, len(v))
			for i, value := range v {
				cids[i] = value.(string)
			}
			commit.CIDs = immutable.Some(cids)

		case request.OrderClause:
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}
			conditions, err := parseOrderConditionList(v)
			if err != nil {
				return nil, err
			}
			commit.OrderBy = immutable.Some(request.OrderBy{
				Conditions: conditions,
			})

		case request.LimitClause:
			if v, ok := value.(int32); ok {
				commit.Limit = immutable.Some(uint64(v))
			}

		case request.OffsetClause:
			if v, ok := value.(int32); ok {
				commit.Offset = immutable.Some(uint64(v))
			}

		case request.DepthClause:
			if v, ok := value.(int32); ok {
				commit.Depth = immutable.Some(uint64(v))
			}

		case request.GroupByClause:
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}
			fields := make([]string, len(v))
			for i, c := range v {
				fields[i] = c.(string)
			}
			commit.GroupBy = immutable.Some(request.GroupBy{
				Fields: fields,
			})

		case request.FilterClause:
			if v, ok := value.(map[string]any); ok {
				commit.Filter = immutable.Some(request.Filter{Conditions: v})
			}
		}
	}

	// no sub fields (unlikely)
	if len(field.selectionSet) == 0 {
		return commit, nil
	}

	fields, err := parseSelectFields(exe, field.selectionSet)
	if err != nil {
		return nil, err
	}
	commit.Fields = fields

	return commit, err
}
