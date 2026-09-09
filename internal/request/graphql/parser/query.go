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
	"strings"

	"github.com/sourcenetwork/immutable"
	wgast "github.com/wundergraph/graphql-go-tools/v2/pkg/ast"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema/types"
)

// parseQueryOperationDefinition parses the individual GraphQL
// 'query' operations, which there may be multiple of.
func parseQueryOperationDefinition(
	exe *executionContext,
	collectedFields [][]*field,
) (*request.OperationDefinition, []error) {
	var selections []request.Selection
	for _, fields := range collectedFields {
		for _, field := range fields {
			var parsedSelection request.Selection
			if field.name == request.CommitsName {
				parsed, err := parseCommitSelect(exe, field)
				if err != nil {
					return nil, []error{err}
				}

				parsedSelection = parsed
			} else if _, isAggregate := request.Aggregates[field.name]; isAggregate {
				parsed, err := parseAggregate(exe, field)
				if err != nil {
					return nil, []error{err}
				}

				// Top-level aggregates must be wrapped in a top-level Select for now
				parsedSelection = &request.Select{
					Field: request.Field{
						Name:  parsed.Name,
						Alias: parsed.Alias,
					},
					ChildSelect: request.ChildSelect{
						Fields: []request.Selection{
							parsed,
						},
					},
				}
			} else {
				// the query doesn't match a reserve name
				// so its probably a generated query
				parsed, err := parseSelect(exe, field)
				if err != nil {
					return nil, []error{err}
				}

				errors := parsed.Validate()
				if len(errors) > 0 {
					return nil, errors
				}

				parsedSelection = parsed
			}
			selections = append(selections, parsedSelection)
		}
	}

	return &request.OperationDefinition{
		Selections: selections,
	}, nil
}

// @todo: Create separate select parse functions
// for generated object queries, and general
// API queries

// parseSelect parses a typed selection field
// which includes sub fields, and may include
// filters, limits, orders, etc..
func parseSelect(
	exe *executionContext,
	field *field,
) (*request.Select, error) {
	isEncrypted := strings.HasPrefix(field.name, request.EncryptedCollectionPrefix)

	slct := &request.Select{
		Field: request.Field{
			Name:  field.name,
			Alias: field.alias,
		},
		IsEncrypted: isEncrypted,
	}

	for _, name := range field.argumentOrder {
		value := field.arguments[name]

		switch name {
		case request.FilterClause:
			if v, ok := value.(map[string]any); ok {
				slct.Filter = immutable.Some(request.Filter{Conditions: v})
			}

		case request.DocIDArgName: // parse single DocID field
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}
			docIDs := make([]string, len(v))
			for i, value := range v {
				docIDs[i] = value.(string)
			}
			slct.DocIDs = immutable.Some(docIDs)

		case request.CidFieldName: // parse single CID query field
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}

			cids := make([]string, len(v))
			for i, value := range v {
				cids[i] = value.(string)
			}
			slct.CIDs = immutable.Some(cids)

		case request.LimitClause: // parse limit/offset
			if v, ok := value.(int32); ok {
				slct.Limit = immutable.Some(uint64(v))
			}

		case request.OffsetClause: // parse limit/offset
			if v, ok := value.(int32); ok {
				slct.Offset = immutable.Some(uint64(v))
			}

		case request.OrderClause: // parse order by
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}
			conditions, err := parseOrderConditionList(v)
			if err != nil {
				return nil, err
			}
			slct.OrderBy = immutable.Some(request.OrderBy{
				Conditions: conditions,
			})

		case request.GroupByClause:
			v, ok := value.([]any)
			if !ok {
				continue // value is nil
			}
			fields := make([]string, len(v))
			for i, c := range v {
				fields[i] = c.(string)
			}
			slct.GroupBy = immutable.Some(request.GroupBy{
				Fields: fields,
			})

		case request.ShowDeleted:
			if v, ok := value.(bool); ok {
				slct.ShowDeleted = v
			}
		}
	}

	// if theres no field selections, just return
	if len(field.selectionSet) == 0 {
		return slct, nil
	}

	// parse field selections
	selections, err := parseSelectFields(exe, field.selectionSet)
	if err != nil {
		return nil, err
	}
	slct.Fields = selections

	return slct, err
}

func parseAggregate(
	exe *executionContext,
	field *field,
) (*request.Aggregate, error) {
	var targets []*request.AggregateTarget
	for _, name := range field.argumentOrder {

		switch v := field.arguments[name].(type) {
		case string:
			targets = append(targets, &request.AggregateTarget{
				HostName: v,
			})

		case map[string]any:
			target, err := parseAggregateTarget(name, v)
			if err != nil {
				return nil, err
			}
			targets = append(targets, target)
		}
	}

	return &request.Aggregate{
		Field: request.Field{
			Name:  field.name,
			Alias: field.alias,
		},
		Targets: targets,
	}, nil
}

func parseSimilarity(
	exe *executionContext,
	field *field,
) (*request.Similarity, error) {
	var target string
	var vector any
	for _, name := range field.argumentOrder {
		target = name
		v := field.arguments[target].(map[string]any)
		vector = v[types.SimilarityArgVector]
	}
	// The argument names the field to compare against, so without one there is nothing to
	// compare. The mapper looks the target up by name and would panic on the empty name.
	if target == "" {
		return nil, ErrSimilarityMissingTarget
	}

	return &request.Similarity{
		Field: request.Field{
			Name:  field.name,
			Alias: field.alias,
		},
		Target: target,
		Vector: vector,
	}, nil
}

func parseAggregateTarget(
	hostName string,
	arguments map[string]any,
) (*request.AggregateTarget, error) {
	var childName string
	var filter immutable.Option[request.Filter]
	var limit immutable.Option[uint64]
	var offset immutable.Option[uint64]
	var order immutable.Option[request.OrderBy]
	var groupBy immutable.Option[request.GroupBy]

	for name, value := range arguments {
		switch name {
		case request.FieldArgName:
			if v, ok := value.(string); ok {
				childName = v
			}

		case request.FilterClause:
			if v, ok := value.(map[string]any); ok {
				filter = immutable.Some(request.Filter{Conditions: v})
			}

		case request.LimitClause:
			if v, ok := value.(int32); ok {
				limit = immutable.Some(uint64(v))
			}

		case request.OffsetClause:
			if v, ok := value.(int32); ok {
				offset = immutable.Some(uint64(v))
			}

		case request.OrderClause:
			switch t := value.(type) {
			case string:
				dir, err := parseOrderDirectionString(t)
				if err != nil {
					return nil, err
				}
				order = immutable.Some(request.OrderBy{
					Conditions: []request.OrderCondition{{Direction: dir}},
				})

			case int:
				// For inline arrays the order arg will be a simple enum declaring the order direction
				dir, err := parseOrderDirection(t)
				if err != nil {
					return nil, err
				}
				order = immutable.Some(request.OrderBy{
					Conditions: []request.OrderCondition{{Direction: dir}},
				})

			case []any:
				// For relations the order arg will be the complex order object as used by the host object
				// for non-aggregate ordering
				conditions, err := parseOrderConditionList(t)
				if err != nil {
					return nil, err
				}
				order = immutable.Some(request.OrderBy{
					Conditions: conditions,
				})
			}

		case request.GroupByClause:
			if raw, ok := value.([]any); ok {
				fields := make([]string, len(raw))
				for i, f := range raw {
					if s, ok := f.(string); ok {
						fields[i] = s
					}
				}
				groupBy = immutable.Some(request.GroupBy{Fields: fields})
			}
		}
	}

	return &request.AggregateTarget{
		HostName:  hostName,
		ChildName: immutable.Some(childName),
		Filterable: request.Filterable{
			Filter: filter,
		},
		Limitable: request.Limitable{
			Limit: limit,
		},
		Offsetable: request.Offsetable{
			Offset: offset,
		},
		Orderable: request.Orderable{
			OrderBy: order,
		},
		Groupable: request.Groupable{
			GroupBy: groupBy,
		},
	}, nil
}

// ValidateSimilarityArgs reports similarity arguments naming a field that cannot hold a vector.
// Such a field has no similarity argument, so the GraphQL library calls it an unknown argument,
// which reads as if the field did not exist. Must run before that validation rejects the request.
func ValidateSimilarityArgs(definition, doc *wgast.Document) []error {
	fragments := make(map[string]int, len(doc.FragmentDefinitions))
	for ref := range doc.FragmentDefinitions {
		fragments[doc.FragmentDefinitionNameString(ref)] = doc.FragmentDefinitions[ref].SelectionSet
	}
	var errs []error
	for _, operation := range doc.OperationDefinitions {
		// Similarity exists on every object type, so a mutation's result set can select it too.
		root, err := operationRootType(definition, operation.OperationType)
		if err != nil {
			continue
		}
		errs = append(errs, validateSimilarityArgs(
			doc, definition, root, operation.SelectionSet, fragments, map[string]bool{})...)
	}
	return errs
}

// validateSimilarityArgs checks obj's similarity selections, then recurses into the related objects
// selected alongside them. visited guards against a fragment cycle.
func validateSimilarityArgs(
	doc *wgast.Document,
	definition *wgast.Document,
	obj wgast.Node,
	selectionSet int,
	fragments map[string]int,
	visited map[string]bool,
) []error {
	similarityArgs := map[string]struct{}{}
	if similarity, exists := definition.NodeFieldDefinitionByName(obj, []byte(request.SimilarityFieldName)); exists {
		for _, arg := range definition.FieldDefinitionArgumentsDefinitions(similarity) {
			similarityArgs[definition.InputValueDefinitionNameString(arg)] = struct{}{}
		}
	}

	var errs []error
	for _, selectionRef := range doc.SelectionSets[selectionSet].SelectionRefs {
		selection := doc.Selections[selectionRef]
		switch selection.Kind {
		case wgast.SelectionKindInlineFragment:
			errs = append(errs, validateSimilarityArgs(
				doc, definition, obj, doc.InlineFragments[selection.Ref].SelectionSet, fragments, visited)...)

		case wgast.SelectionKindFragmentSpread:
			name := doc.FragmentSpreadNameString(selection.Ref)
			if visited[name] {
				continue
			}
			visited[name] = true
			if fragment, exists := fragments[name]; exists {
				errs = append(errs, validateSimilarityArgs(doc, definition, obj, fragment, fragments, visited)...)
			}

		case wgast.SelectionKindField:
			name := doc.FieldNameString(selection.Ref)
			if name != request.SimilarityFieldName {
				if !doc.Fields[selection.Ref].HasSelections {
					continue
				}
				errs = append(errs, validateSimilarityArgs(
					doc,
					definition,
					objectOf(definition, obj, name),
					doc.Fields[selection.Ref].SelectionSet,
					fragments,
					visited,
				)...)
				continue
			}
			errs = append(errs, validateSimilarityFieldArgs(
				doc, definition, obj, selection.Ref, similarityArgs)...)
		}
	}
	return errs
}

func validateSimilarityFieldArgs(
	doc *wgast.Document,
	definition *wgast.Document,
	obj wgast.Node,
	similarity int,
	similarityArgs map[string]struct{},
) []error {
	var errs []error
	for _, arg := range doc.Fields[similarity].Arguments.Refs {
		name := doc.ArgumentNameString(arg)
		if _, isSimilarityArg := similarityArgs[name]; isSimilarityArg {
			continue
		}
		field, exists := definition.NodeFieldDefinitionByName(obj, []byte(name))
		if !exists {
			// Not a field at all, so the library's error is already right.
			continue
		}
		fieldType, err := definition.PrintTypeBytes(definition.FieldDefinitionType(field), nil)
		if err != nil {
			continue
		}
		errs = append(errs, NewErrSimilarityOnNonVectorField(name, string(fieldType)))
	}
	return errs
}

// objectOf resolves the object type behind a field, through the list and non-null wrappers a
// collection or relation field is built from.
func objectOf(definition *wgast.Document, obj wgast.Node, fieldName string) wgast.Node {
	field, exists := definition.NodeFieldDefinitionByName(obj, []byte(fieldName))
	if !exists {
		return wgast.Node{}
	}
	return definition.FieldDefinitionTypeNode(field)
}
