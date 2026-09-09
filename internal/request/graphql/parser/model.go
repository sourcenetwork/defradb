// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.

package parser

import (
	"encoding/json"
	"fmt"

	"github.com/sourcenetwork/immutable"
	wgast "github.com/wundergraph/graphql-go-tools/v2/pkg/ast"

	"github.com/sourcenetwork/defradb/client/request"
)

type operation struct {
	typ        wgast.OperationType
	directives request.Directives
	fields     [][]*field
}

type field struct {
	name          string
	alias         immutable.Option[string]
	argumentOrder []string
	arguments     map[string]any
	selectionSet  []*field
}

func buildOperation(definition, document *wgast.Document) (*operation, error) {
	for _, root := range document.RootNodes {
		if root.Kind != wgast.NodeKindOperationDefinition {
			continue
		}
		operationDefinition := document.OperationDefinitions[root.Ref]
		parent, err := operationRootType(definition, operationDefinition.OperationType)
		if err != nil {
			return nil, err
		}
		fields, err := buildFields(document, definition, parent, operationDefinition.SelectionSet)
		if err != nil {
			return nil, err
		}
		directives, err := buildDirectives(document, operationDefinition.Directives.Refs)
		if err != nil {
			return nil, err
		}
		return &operation{
			typ:        operationDefinition.OperationType,
			directives: directives,
			fields:     groupFields(fields),
		}, nil
	}
	return nil, fmt.Errorf("normalized request does not contain an operation")
}

func operationRootType(definition *wgast.Document, operationType wgast.OperationType) (wgast.Node, error) {
	var name []byte
	switch operationType {
	case wgast.OperationTypeQuery:
		name = definition.Index.QueryTypeName
	case wgast.OperationTypeMutation:
		name = definition.Index.MutationTypeName
	case wgast.OperationTypeSubscription:
		name = definition.Index.SubscriptionTypeName
	default:
		return wgast.Node{}, ErrUnknownGQLOperation
	}
	if len(name) == 0 {
		return wgast.Node{}, fmt.Errorf("schema is not configured for %s operations", operationType)
	}
	root, exists := definition.NodeByName(name)
	if !exists {
		return wgast.Node{}, fmt.Errorf("schema root type %s is not defined", name)
	}
	return root, nil
}

func buildFields(
	document *wgast.Document,
	definition *wgast.Document,
	parent wgast.Node,
	selectionSet int,
) ([]*field, error) {
	var result []*field
	for _, selectionRef := range document.SelectionSets[selectionSet].SelectionRefs {
		selection := document.Selections[selectionRef]
		switch selection.Kind {
		case wgast.SelectionKindInlineFragment:
			fields, err := buildFields(document, definition, parent, document.InlineFragments[selection.Ref].SelectionSet)
			if err != nil {
				return nil, err
			}
			result = append(result, fields...)
		case wgast.SelectionKindFragmentSpread:
			return nil, fmt.Errorf("normalized request contains a fragment spread")
		case wgast.SelectionKindField:
			built, err := buildField(document, definition, parent, selection.Ref)
			if err != nil {
				return nil, err
			}
			result = append(result, built)
		}
	}
	return result, nil
}

func buildField(document, definition *wgast.Document, parent wgast.Node, ref int) (*field, error) {
	name := document.FieldNameString(ref)
	if name == request.TypeNameFieldName {
		result := &field{name: name}
		if document.FieldAliasIsDefined(ref) {
			result.alias = immutable.Some(document.FieldAliasString(ref))
		}
		return result, nil
	}
	fieldDefinition, exists := definition.NodeFieldDefinitionByName(parent, []byte(name))
	if !exists {
		return nil, fmt.Errorf("field %s is not defined", name)
	}
	arguments, order, err := buildArguments(
		document,
		definition,
		definition.FieldDefinitionArgumentsDefinitions(fieldDefinition),
		document.Fields[ref].Arguments.Refs,
	)
	if err != nil {
		return nil, err
	}
	result := &field{
		name:          name,
		argumentOrder: order,
		arguments:     arguments,
	}
	if document.FieldAliasIsDefined(ref) {
		result.alias = immutable.Some(document.FieldAliasString(ref))
	}
	if document.Fields[ref].HasSelections {
		childType := definition.FieldDefinitionTypeNode(fieldDefinition)
		result.selectionSet, err = buildFields(document, definition, childType, document.Fields[ref].SelectionSet)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func buildArguments(
	document *wgast.Document,
	definition *wgast.Document,
	definitions []int,
	refs []int,
) (map[string]any, []string, error) {
	definitionsByName := make(map[string]int, len(definitions))
	result := make(map[string]any, len(definitions))
	for _, definitionRef := range definitions {
		name := definition.InputValueDefinitionNameString(definitionRef)
		definitionsByName[name] = definitionRef
		if definition.InputValueDefinitionHasDefaultValue(definitionRef) {
			valueJSON, err := definition.ValueToJSON(definition.InputValueDefinitionDefaultValue(definitionRef))
			if err != nil {
				return nil, nil, err
			}
			var value any
			if err := json.Unmarshal(valueJSON, &value); err != nil {
				return nil, nil, err
			}
			result[name] = coerceInputValue(definition, definition.InputValueDefinitionType(definitionRef), value)
		}
	}
	order := make([]string, 0, len(refs))
	for _, ref := range refs {
		name := document.ArgumentNameString(ref)
		argumentDefinition, exists := definitionsByName[name]
		if !exists {
			return nil, nil, fmt.Errorf("argument %s is not defined", name)
		}
		order = append(order, name)
		value := document.ArgumentValue(ref)
		if value.Kind == wgast.ValueKindVariable {
			variableName := document.VariableValueNameString(value.Ref)
			var providedVariables map[string]json.RawMessage
			if err := json.Unmarshal(document.Input.Variables, &providedVariables); err != nil {
				return nil, nil, err
			}
			if _, exists := providedVariables[variableName]; !exists {
				continue
			}
		}
		valueJSON, err := document.ValueToJSON(value)
		if err != nil {
			return nil, nil, err
		}
		var decodedValue any
		if err := json.Unmarshal(valueJSON, &decodedValue); err != nil {
			return nil, nil, err
		}
		typeRef := definition.InputValueDefinitionType(argumentDefinition)
		if err := validateInputValue(definition, typeRef, decodedValue); err != nil {
			return nil, nil, err
		}
		result[name] = coerceInputValue(definition, typeRef, decodedValue)
	}
	return result, order, nil
}

func groupFields(fields []*field) [][]*field {
	groups := make(map[string][]*field)
	order := make([]string, 0, len(fields))
	for _, field := range fields {
		key := field.name
		if field.alias.HasValue() {
			key = field.alias.Value()
		}
		if _, exists := groups[key]; !exists {
			order = append(order, key)
		}
		groups[key] = append(groups[key], field)
	}
	result := make([][]*field, 0, len(groups))
	for _, key := range order {
		result = append(result, groups[key])
	}
	return result
}

func buildDirectives(document *wgast.Document, refs []int) (request.Directives, error) {
	result := request.Directives{
		ExplainType: immutable.None[request.ExplainType](),
	}
	for _, ref := range refs {
		switch document.DirectiveNameString(ref) {
		case request.ExhaustiveLabel:
			result.Exhaustive = true
		case request.ExplainLabel:
			result.ExplainType = immutable.Some(request.SimpleExplain)
			arguments := document.Directives[ref].Arguments.Refs
			if len(arguments) == 0 {
				continue
			}
			if len(arguments) != 1 {
				return request.Directives{}, ErrInvalidNumberOfExplainArgs
			}
			if document.ArgumentNameString(arguments[0]) != "type" {
				return request.Directives{}, ErrInvalidExplainTypeArg
			}
			valueJSON, err := document.ValueToJSON(document.ArgumentValue(arguments[0]))
			if err != nil {
				return request.Directives{}, err
			}
			var explainType string
			if err := json.Unmarshal(valueJSON, &explainType); err != nil {
				return request.Directives{}, err
			}
			switch explainType {
			case "simple":
				result.ExplainType = immutable.Some(request.SimpleExplain)
			case "execute":
				result.ExplainType = immutable.Some(request.ExecuteExplain)
			case "debug":
				result.ExplainType = immutable.Some(request.DebugExplain)
			default:
				return request.Directives{}, ErrUnknownExplainType
			}
		}
	}
	return result, nil
}
