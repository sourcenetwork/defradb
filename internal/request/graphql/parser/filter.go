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
	"fmt"
	"strings"

	wgast "github.com/wundergraph/graphql-go-tools/v2/pkg/ast"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astparser"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

// NewFilterFromString creates a new filter from a string.
func NewFilterFromString(
	definition *wgast.Document,
	collectionType string,
	body string,
) (immutable.Option[request.Filter], error) {
	if !strings.HasPrefix(body, "{") {
		body = "{" + body + "}"
	}
	document, report := astparser.ParseGraphqlDocumentString(fmt.Sprintf(
		"{ %s(%s: %s) }",
		collectionType,
		request.FilterClause,
		body,
	))
	if report.HasErrors() {
		return immutable.None[request.Filter](), report
	}
	operation, err := buildOperation(definition, &document)
	if err != nil {
		return immutable.None[request.Filter](), err
	}
	if len(operation.fields) == 0 || len(operation.fields[0]) == 0 {
		return immutable.None[request.Filter](), ErrFilterMissingArgumentType
	}
	conditions, ok := operation.fields[0][0].arguments[request.FilterClause].(map[string]any)
	if !ok {
		return immutable.None[request.Filter](), ErrFailedToParseConditionsFromAST
	}
	return immutable.Some(request.Filter{Conditions: conditions}), nil
}

// ParseFilterFieldsForDescription parses the fields that are defined in the SchemaDescription
// from the filter conditions“
func ParseFilterFieldsForDescription(
	conditions map[string]any,
	col client.CollectionVersion,
) ([]client.CollectionFieldDescription, error) {
	return parseFilterFieldsForDescriptionMap(conditions, col)
}

func parseFilterFieldsForDescriptionMap(
	conditions map[string]any,
	col client.CollectionVersion,
) ([]client.CollectionFieldDescription, error) {
	fields := make([]client.CollectionFieldDescription, 0)
	for k, v := range conditions {
		switch k {
		case request.FilterOpOr, request.FilterOpAnd:
			conds := v.([]any)
			parsedFields, err := parseFilterFieldsForDescriptionSlice(conds, col)
			if err != nil {
				return nil, err
			}
			fields = append(fields, parsedFields...)
		case request.FilterOpNot, request.AliasFieldName:
			conds := v.(map[string]any)
			parsedFields, err := parseFilterFieldsForDescriptionMap(conds, col)
			if err != nil {
				return nil, err
			}
			fields = append(fields, parsedFields...)
		default:
			f, found := col.GetFieldByName(k)
			if !found || f.Kind.IsObject() {
				continue
			}
			fields = append(fields, f)
		}
	}
	return fields, nil
}

func parseFilterFieldsForDescriptionSlice(
	conditions []any,
	schema client.CollectionVersion,
) ([]client.CollectionFieldDescription, error) {
	fields := make([]client.CollectionFieldDescription, 0)
	for _, v := range conditions {
		switch cond := v.(type) {
		case map[string]any:
			parsedFields, err := parseFilterFieldsForDescriptionMap(cond, schema)
			if err != nil {
				return nil, err
			}
			fields = append(fields, parsedFields...)
		default:
			return nil, ErrInvalidFilterConditions
		}
	}
	return fields, nil
}
