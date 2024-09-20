// Copyright 2024 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/client/request"
)

func parseOrderConditionList(args []any) ([]request.OrderCondition, error) {
	var conditions []request.OrderCondition
	for _, a := range args {
		v, ok := a.(map[string]any)
		if !ok {
			continue // order value is nil
		}
		condition, err := parseOrderCondition(v)
		if err != nil {
			return nil, err
		}
		if condition != nil {
			conditions = append(conditions, *condition)
		}
	}
	return conditions, nil
}

func parseOrderCondition(arg map[string]any) (*request.OrderCondition, error) {
	if len(arg) == 0 {
		return nil, nil
	}
	if len(arg) != 1 {
		return nil, ErrMultipleOrderFieldsDefined
	}
	var fieldName string
	for name := range arg {
		fieldName = name
	}
	switch t := arg[fieldName].(type) {
	case int:
		dir, err := parseOrderDirection(t)
		if err != nil {
			return nil, err
		}
		return &request.OrderCondition{
			Fields:    []string{fieldName},
			Direction: dir,
		}, nil

	case map[string]any:
		cond, err := parseOrderCondition(t)
		if err != nil {
			return nil, err
		}
		if cond == nil {
			return nil, nil
		}
		// prepend the current field name, to the parsed condition from the slice
		// Eg. order: {author: {name: ASC, birthday: DESC}}
		// This results in an array of [name, birthday] converted to
		// [author.name, author.birthday].
		// etc.
		cond.Fields = append([]string{fieldName}, cond.Fields...)
		return cond, nil

	default:
		// field value is null so don't include the condition
		return nil, nil
	}
}

func parseOrderDirection(v int) (request.OrderDirection, error) {
	switch v {
	case 0:
		return request.ASC, nil

	case 1:
		return request.DESC, nil

	default:
		return request.ASC, ErrInvalidOrderDirection
	}
}
