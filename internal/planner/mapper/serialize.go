// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mapper

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/core"
)

type connorJSONSerializer struct{}

func (s connorJSONSerializer) serializeConditions(conditions map[connor.FilterKey]any) ([]byte, error) {
	return json.Marshal(s.processCondition(conditions))
}

func (s connorJSONSerializer) processCondition(conditions any) map[string]any {
	serEl := func(key connor.FilterKey, value any) map[string]any {
		if op, ok := key.(*Operator); ok {
			return map[string]any{strings.ToUpper(op.Operation[1:]): s.processOp(op.Operation, value)}
		} else if prop, ok := key.(*PropertyIndex); ok {
			return map[string]any{"PROP": s.processProp(prop, value)}
		}
		return nil
	}

	if m, ok := conditions.(map[connor.FilterKey]any); ok {
		if len(m) > 1 {
			andCont := make([]any, 0, len(m))
			for key, value := range m {
				andCont = append(andCont, serEl(key, value))
			}
			return map[string]any{"AND": andCont}
		} else {
			for key, value := range m {
				return serEl(key, value)
			}
		}
	}
	return nil
}

func (s connorJSONSerializer) processProp(prop *PropertyIndex, val any) map[string]any {
	return map[string]any{
		"index":     prop.Index,
		"condition": s.processCondition(val),
	}
}

func (s connorJSONSerializer) processOp(op string, val any) any {
	switch op {
	case "_and", "_or":
		if arr, ok := val.([]any); ok {
			arrResult := make([]any, len(arr))
			for i, v := range arr {
				arrResult[i] = s.processCondition(v)
			}
			return arrResult
		}
	case "_not":
		return s.processCondition(val)
	}
	return s.processField(val)
}

func (s connorJSONSerializer) processField(val any) map[string]any {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case bool:
		return map[string]any{"Bool": v}
	case int32, int64:
		return map[string]any{"Int": v}
	case float32, float64:
		return map[string]any{"Float": v}
	case string:
		return map[string]any{"String": v}
	case time.Time:
		return map[string]any{"DateTime": v}
	case core.Doc:
		return map[string]any{"Doc": s.processDoc(v)}
	case []bool:
		return map[string]any{"BoolArray": v}
	case []int32, []int64:
		return map[string]any{"IntArray": v}
	case []float32, []float64:
		return map[string]any{"FloatArray": v}
	case []string:
		return map[string]any{"StringArray": v}
	case []time.Time:
		return map[string]any{"DateTimeArray": v}
	case []core.Doc:
		return map[string]any{"DocArray": s.processDocArray(v)}
	case immutable.Option[bool]:
		return unwrapOption("Bool", v)
	case immutable.Option[int64]:
		return unwrapOption("Int", v)
	case immutable.Option[float64]:
		return unwrapOption("Float", v)
	case immutable.Option[string]:
		return unwrapOption("String", v)
	case immutable.Option[time.Time]:
		return unwrapOption("DateTime", v)
	case []any:
		return s.processAnyArr(v)
	default:
		return nil
	}
}

func unwrapOption[T any](t string, opt immutable.Option[T]) map[string]any {
	if opt.HasValue() {
		return map[string]any{t: opt.Value()}
	}
	return nil
}

func (s connorJSONSerializer) processAnyArr(arr []any) map[string]any {
	if len(arr) == 0 {
		return nil
	}
	var el any
	for _, v := range arr {
		if v != nil {
			el = v
			break
		}
	}
	switch el.(type) {
	case bool:
		return s.toArrayOrOptArray("BoolArray", arr)
	case int32, int64:
		return s.toArrayOrOptArray("IntArray", arr)
	case float32, float64:
		return s.toArrayOrOptArray("FloatArray", arr)
	case string:
		return s.toArrayOrOptArray("StringArray", arr)
	case time.Time:
		return s.toArrayOrOptArray("DateTimeArray", arr)
	case core.Doc:
		result := make([]map[string]any, len(arr))
		for i, doc := range arr {
			result[i] = s.processDoc(doc.(core.Doc))
		}
		return map[string]any{"DocArray": result}
	default:
		return nil
	}
}

func (s connorJSONSerializer) toArrayOrOptArray(fieldType string, arr []any) map[string]any {
	for _, v := range arr {
		if v == nil {
			return map[string]any{"Optional" + fieldType: arr}
		}
	}
	return map[string]any{fieldType: arr}
}

func (s connorJSONSerializer) processDocArray(docs []core.Doc) []map[string]any {
	result := make([]map[string]any, len(docs))
	for i, doc := range docs {
		result[i] = s.processDoc(doc)
	}
	return result
}

func (s connorJSONSerializer) serializeDoc(doc core.Doc) ([]byte, error) {
	return json.Marshal(s.processDoc(doc))
}

func (s connorJSONSerializer) serializeDocField(field any) ([]byte, error) {
	return json.Marshal(s.processField(field))
}

func (s connorJSONSerializer) processDoc(doc core.Doc) map[string]any {
	fields := make([]any, len(doc.Fields))
	for i, field := range doc.Fields {
		fields[i] = s.processField(field)
	}
	return map[string]any{"fields": fields}
}
