// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.

package parser

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	wgast "github.com/wundergraph/graphql-go-tools/v2/pkg/ast"
)

type utcNowValue struct{}

func validateInputValue(definition *wgast.Document, typeRef int, value any) error {
	typeDefinition := definition.Types[typeRef]
	if value == nil {
		if typeDefinition.TypeKind == wgast.TypeKindNonNull {
			printed, err := definition.PrintTypeBytes(typeRef, nil)
			if err != nil {
				return err
			}
			return fmt.Errorf("Expected %q, found null.", string(printed))
		}
		return nil
	}
	switch typeDefinition.TypeKind {
	case wgast.TypeKindNonNull:
		return validateInputValue(definition, typeDefinition.OfType, value)
	case wgast.TypeKindList:
		valueType := reflect.ValueOf(value)
		if valueType.Kind() != reflect.Slice && valueType.Kind() != reflect.Array {
			return validateInputValue(definition, typeDefinition.OfType, value)
		}
		for index := range valueType.Len() {
			if err := validateInputValue(definition, typeDefinition.OfType, valueType.Index(index).Interface()); err != nil {
				return err
			}
		}
	case wgast.TypeKindNamed:
		name := definition.TypeNameString(typeRef)
		node, exists := definition.NodeByNameStr(name)
		input, ok := value.(map[string]any)
		if !exists || node.Kind != wgast.NodeKindInputObjectTypeDefinition || !ok {
			return nil
		}
		for _, fieldRef := range definition.NodeInputFieldDefinitions(node) {
			fieldName := definition.InputValueDefinitionNameString(fieldRef)
			if fieldValue, exists := input[fieldName]; exists {
				if err := validateInputValue(definition, definition.InputValueDefinitionType(fieldRef), fieldValue); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func coerceInputValue(definition *wgast.Document, typeRef int, value any) any {
	if value == nil {
		return nil
	}
	typeDefinition := definition.Types[typeRef]
	switch typeDefinition.TypeKind {
	case wgast.TypeKindNonNull:
		return coerceInputValue(definition, typeDefinition.OfType, value)
	case wgast.TypeKindList:
		valueType := reflect.ValueOf(value)
		if valueType.Kind() != reflect.Slice && valueType.Kind() != reflect.Array {
			return []any{coerceInputValue(definition, typeDefinition.OfType, value)}
		}
		result := make([]any, valueType.Len())
		for index := range valueType.Len() {
			result[index] = coerceInputValue(definition, typeDefinition.OfType, valueType.Index(index).Interface())
		}
		return result
	case wgast.TypeKindNamed:
		name := definition.TypeNameString(typeRef)
		node, exists := definition.NodeByNameStr(name)
		if exists && node.Kind == wgast.NodeKindInputObjectTypeDefinition {
			return coerceInputObject(definition, node, value)
		}
		return coerceScalar(name, value)
	default:
		return value
	}
}

func coerceInputObject(definition *wgast.Document, node wgast.Node, value any) any {
	input, ok := value.(map[string]any)
	if !ok {
		return value
	}
	result := make(map[string]any, len(input))
	for _, fieldRef := range definition.NodeInputFieldDefinitions(node) {
		name := definition.InputValueDefinitionNameString(fieldRef)
		fieldValue, exists := input[name]
		switch {
		case exists:
			result[name] = coerceInputValue(definition, definition.InputValueDefinitionType(fieldRef), fieldValue)
		case definition.InputValueDefinitionHasDefaultValue(fieldRef):
			valueJSON, err := definition.ValueToJSON(definition.InputValueDefinitionDefaultValue(fieldRef))
			if err != nil {
				continue
			}
			var defaultValue any
			if json.Unmarshal(valueJSON, &defaultValue) == nil {
				result[name] = coerceInputValue(definition, definition.InputValueDefinitionType(fieldRef), defaultValue)
			}
		}
	}
	return result
}

func coerceScalar(name string, value any) any {
	switch name {
	case "Int":
		switch value := value.(type) {
		case float64:
			return int32(value)
		case json.Number:
			parsed, err := strconv.ParseInt(value.String(), 10, 32)
			if err == nil {
				return int32(parsed)
			}
		}
	case "Float32":
		if value, ok := value.(float64); ok {
			return float32(value)
		}
	case "DateTime":
		if value, ok := value.(string); ok {
			if value == "UTC_NOW" {
				return utcNowValue{}
			}
			if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
				return parsed
			}
		}
	case "JSON":
		return coerceJSONNumbers(value)
	}
	return value
}

func coerceJSONNumbers(value any) any {
	switch value := value.(type) {
	case float64:
		integer := int32(value)
		if float64(integer) == value {
			return integer
		}
	case []any:
		result := make([]any, len(value))
		for index, item := range value {
			result[index] = coerceJSONNumbers(item)
		}
		return result
	case map[string]any:
		result := make(map[string]any, len(value))
		for name, item := range value {
			result[name] = coerceJSONNumbers(item)
		}
		return result
	}
	return value
}
