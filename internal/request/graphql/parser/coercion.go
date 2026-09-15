// Copyright 2026 Democratized Data Foundation
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
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	wgast "github.com/wundergraph/graphql-go-tools/v2/pkg/ast"

	"github.com/sourcenetwork/defradb/internal/request/graphql/schema/types"
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
			return newErrExpectedNull(string(printed))
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
		if !exists {
			return nil
		}
		if node.Kind != wgast.NodeKindInputObjectTypeDefinition {
			return validateScalar(name, value)
		}
		if !ok {
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

func validateScalar(name string, value any) error {
	switch name {
	case "Int":
		if number, ok := numericValue(value); !ok || math.Trunc(number) != number ||
			number < math.MinInt32 || number > math.MaxInt32 {
			return fmt.Errorf("Int cannot represent non 32-bit signed integer value: %v", value)
		}
	case "DateTime":
		text, ok := value.(string)
		if !ok {
			return fmt.Errorf("DateTime cannot represent value: %v", value)
		}
		if text != "UTC_NOW" {
			if _, err := time.Parse(time.RFC3339Nano, text); err != nil {
				return fmt.Errorf("DateTime cannot represent value: %q", text)
			}
		}
	case "Blob":
		text, ok := value.(string)
		if !ok || !types.BlobPattern.MatchString(text) {
			return fmt.Errorf("Blob cannot represent value: %v", value)
		}
	}
	return nil
}

func numericValue(value any) (float64, bool) {
	switch value := value.(type) {
	case float64:
		return value, true
	case float32:
		return float64(value), true
	case int:
		return float64(value), true
	case int8:
		return float64(value), true
	case int16:
		return float64(value), true
	case int32:
		return float64(value), true
	case int64:
		return float64(value), true
	case uint:
		return float64(value), true
	case uint8:
		return float64(value), true
	case uint16:
		return float64(value), true
	case uint32:
		return float64(value), true
	case uint64:
		return float64(value), true
	case json.Number:
		number, err := strconv.ParseFloat(value.String(), 64)
		return number, err == nil
	default:
		return 0, false
	}
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
