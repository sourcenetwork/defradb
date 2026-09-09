// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.

package graphql

import (
	"encoding/json"
	"fmt"
	"sort"

	wgast "github.com/wundergraph/graphql-go-tools/v2/pkg/ast"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astnormalization"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astparser"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astprinter"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/asttransform"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astvalidation"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/introspection"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/operationreport"

	"github.com/sourcenetwork/defradb/client"
)

func executeIntrospection(definition *wgast.Document, source string) *client.RequestResult {
	reportResult := func(report *operationreport.Report) *client.RequestResult {
		errs := append([]error(nil), report.InternalErrors...)
		for _, gqlErr := range report.ExternalErrors {
			errs = append(errs, fmt.Errorf("%s", gqlErr.Message))
		}
		return introspectionErrorResult(errs)
	}
	validationDefinition, err := introspectionDefinition(definition)
	if err != nil {
		return introspectionErrorResult([]error{err})
	}
	operation, report := astparser.ParseGraphqlDocumentString(source)
	if report.HasErrors() {
		return reportResult(&report)
	}
	astnormalization.NormalizeOperation(&operation, validationDefinition, &report)
	if report.HasErrors() {
		return reportResult(&report)
	}
	if astvalidation.DefaultOperationValidator().Validate(
		&operation,
		validationDefinition,
		&report,
	) == astvalidation.Invalid {
		return reportResult(&report)
	}

	var generated introspection.Data
	introspection.NewGenerator().Generate(definition, &report, &generated)
	if report.HasErrors() {
		return reportResult(&report)
	}
	raw, err := json.Marshal(generated)
	if err != nil {
		return introspectionErrorResult([]error{err})
	}
	var sourceData map[string]any
	if err := json.Unmarshal(raw, &sourceData); err != nil {
		return introspectionErrorResult([]error{err})
	}

	operationRef := -1
	for _, root := range operation.RootNodes {
		if root.Kind == wgast.NodeKindOperationDefinition {
			operationRef = root.Ref
			break
		}
	}
	if operationRef < 0 {
		return introspectionErrorResult([]error{fmt.Errorf("introspection request contains no operation")})
	}
	selectionSet := operation.OperationDefinitions[operationRef].SelectionSet
	data, err := projectIntrospectionSelection(&operation, selectionSet, sourceData, sourceData)
	if err != nil {
		return introspectionErrorResult([]error{err})
	}
	return &client.RequestResult{GQL: client.GQLResult{Data: data}}
}

// introspectionDefinition returns a disposable schema containing GraphQL's
// built-in introspection types and meta-fields. MergeDefinitionWithBaseSchema
// mutates its argument, so the manager's immutable schema snapshot is cloned
// through SDL first.
func introspectionDefinition(definition *wgast.Document) (*wgast.Document, error) {
	sdl, err := astprinter.PrintString(definition)
	if err != nil {
		return nil, err
	}
	clone, report := astparser.ParseGraphqlDocumentString(sdl)
	if report.HasErrors() {
		return nil, report
	}
	if err := asttransform.MergeDefinitionWithBaseSchema(&clone); err != nil {
		return nil, err
	}
	return &clone, nil
}

func projectIntrospectionSelection(
	operation *wgast.Document,
	selectionSet int,
	value any,
	root map[string]any,
) (map[string]any, error) {
	object, _ := value.(map[string]any)
	result := map[string]any{}
	for _, selectionRef := range operation.SelectionSets[selectionSet].SelectionRefs {
		selection := operation.Selections[selectionRef]
		if selection.Kind != wgast.SelectionKindField {
			continue
		}
		fieldRef := selection.Ref
		name := operation.FieldNameString(fieldRef)
		responseName := name
		if operation.FieldAliasIsDefined(fieldRef) {
			responseName = operation.FieldAliasString(fieldRef)
		}

		var fieldValue any
		switch name {
		case "__schema":
			fieldValue = root["__schema"]
		case "__type":
			fieldValue = introspectionTypeByArgument(operation, fieldRef, root)
		default:
			fieldValue = object[name]
			if name == "description" {
				fieldValue = introspectionDescription(fieldValue)
			}
			// The introspection generator emits compact type references for nested
			// __Type values. GraphQL permits clients to traverse those references
			// as full types (for example, arg.type.inputFields), so resolve a
			// compact reference through __schema.types when needed.
			if fieldValue == nil && name == "inputFields" && object["kind"] == "INPUT_OBJECT" {
				if typeName, ok := object["name"].(string); ok {
					if fullType := introspectionTypeByName(typeName, root); fullType != nil {
						fieldValue = fullType[name]
					}
				}
			}
		}
		if name == "args" || name == "fields" || name == "inputFields" {
			fieldValue = sortIntrospectionNamedValues(fieldValue)
		}
		if !operation.Fields[fieldRef].HasSelections || fieldValue == nil {
			result[responseName] = fieldValue
			continue
		}
		projected, err := projectIntrospectionValue(
			operation,
			operation.Fields[fieldRef].SelectionSet,
			fieldValue,
			root,
		)
		if err != nil {
			return nil, err
		}
		result[responseName] = projected
	}
	return result, nil
}

// introspectionDescription converts a description from its SDL source
// representation to its semantic string value. graphql-go-tools currently
// exposes escaped quoted descriptions verbatim (for example, "\\n" instead of
// a newline), unlike the original GraphQL implementation.
func introspectionDescription(value any) any {
	description, ok := value.(string)
	if !ok {
		return value
	}
	var decoded string
	if json.Unmarshal([]byte(`"`+description+`"`), &decoded) == nil {
		return decoded
	}
	return description
}

func sortIntrospectionNamedValues(value any) any {
	values, ok := value.([]any)
	if !ok {
		return value
	}
	values = append([]any(nil), values...)
	sort.SliceStable(values, func(i, j int) bool {
		left, _ := values[i].(map[string]any)["name"].(string)
		right, _ := values[j].(map[string]any)["name"].(string)
		return left < right
	})
	return values
}

func projectIntrospectionValue(
	operation *wgast.Document,
	selectionSet int,
	value any,
	root map[string]any,
) (any, error) {
	switch value := value.(type) {
	case []any:
		result := make([]any, len(value))
		for index, item := range value {
			projected, err := projectIntrospectionSelection(operation, selectionSet, item, root)
			if err != nil {
				return nil, err
			}
			result[index] = projected
		}
		return result, nil
	case map[string]any:
		return projectIntrospectionSelection(operation, selectionSet, value, root)
	default:
		return value, nil
	}
}

func introspectionTypeByArgument(operation *wgast.Document, fieldRef int, root map[string]any) any {
	for _, argumentRef := range operation.Fields[fieldRef].Arguments.Refs {
		if operation.ArgumentNameString(argumentRef) != "name" {
			continue
		}
		value, err := operation.ValueToJSON(operation.ArgumentValue(argumentRef))
		if err != nil {
			return nil
		}
		var name string
		if json.Unmarshal(value, &name) != nil {
			return nil
		}
		if result := introspectionTypeByName(name, root); result != nil {
			return result
		}
		return nil
	}
	return nil
}

func introspectionTypeByName(name string, root map[string]any) map[string]any {
	schema, _ := root["__schema"].(map[string]any)
	types, _ := schema["types"].([]any)
	for _, candidate := range types {
		candidateType, _ := candidate.(map[string]any)
		if candidateType["name"] == name {
			return candidateType
		}
	}
	return nil
}

func introspectionErrorResult(errs []error) *client.RequestResult {
	return &client.RequestResult{GQL: client.GQLResult{Errors: errs}}
}
