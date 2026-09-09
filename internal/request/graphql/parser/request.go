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
	"time"

	wgast "github.com/wundergraph/graphql-go-tools/v2/pkg/ast"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

type executionContext struct {
	now time.Time
}

// ParseRequest translates a normalized graphql-go-tools operation into a
// DefraDB request.
func ParseRequest(definition, document *wgast.Document) (*request.Request, []error) {
	if document == nil {
		return nil, []error{client.NewErrUninitializeProperty("ParseRequest", "document")}
	}
	operation, err := buildOperation(definition, document)
	if err != nil {
		return nil, []error{err}
	}
	exe := &executionContext{now: time.Now().UTC()}

	r := &request.Request{
		Queries:      make([]*request.OperationDefinition, 0),
		Mutations:    make([]*request.OperationDefinition, 0),
		Subscription: make([]*request.OperationDefinition, 0),
	}

	switch operation.typ {
	case wgast.OperationTypeQuery:
		parsedQueryOpDef, errs := parseQueryOperationDefinition(exe, operation.fields)
		if errs != nil {
			return nil, errs
		}
		parsedQueryOpDef.Directives = operation.directives

		r.Queries = append(r.Queries, parsedQueryOpDef)

	case wgast.OperationTypeMutation:
		parsedMutationOpDef, err := parseMutationOperationDefinition(exe, operation.fields)
		if err != nil {
			return nil, []error{err}
		}

		parsedMutationOpDef.Directives = operation.directives

		r.Mutations = append(r.Mutations, parsedMutationOpDef)

	case wgast.OperationTypeSubscription:
		parsedSubscriptionOpDef, errs := parseQueryOperationDefinition(exe, operation.fields)
		if errs != nil {
			return nil, errs
		}

		parsedSubscriptionOpDef.Directives = operation.directives

		r.Subscription = append(r.Subscription, parsedSubscriptionOpDef)

	default:
		return nil, []error{ErrUnknownGQLOperation}
	}

	return r, nil
}

func parseSelectFields(
	exe *executionContext,
	fields []*field,
) ([]request.Selection, error) {
	var selections []request.Selection
	for _, node := range fields {
		var selection request.Selection
		if _, isAggregate := request.Aggregates[node.name]; isAggregate {
			s, err := parseAggregate(exe, node)
			if err != nil {
				return nil, err
			}
			selection = s
		} else if node.name == request.SimilarityFieldName {
			s, err := parseSimilarity(exe, node)
			if err != nil {
				return nil, err
			}
			selection = s
		} else if len(node.selectionSet) == 0 { // regular field
			selection = parseField(node)
		} else if node.name == request.LinksFieldName ||
			node.name == request.HeadsFieldName { // commit links field
			s, err := parseCommitSelect(exe, node)
			if err != nil {
				return nil, err
			}
			selection = s
		} else { // sub type with extra fields
			s, err := parseSelect(exe, node)
			if err != nil {
				return nil, err
			}
			selection = s
		}
		selections = append(selections, selection)
	}

	return selections, nil
}

// parseField simply parses the Name/Alias
// into a Field type
func parseField(field *field) *request.Field {
	return &request.Field{
		Name:  field.name,
		Alias: field.alias,
	}
}
