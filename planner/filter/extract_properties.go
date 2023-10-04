// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package filter

import (
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/planner/mapper"
)

// Property represents a single field and is being filtered on.
// It contains the index of the field in the core.DocumentMapping 
// as well as index -> Property map of the fields in case the field is an object.
type Property struct {
	Index  int
	Fields map[int]Property
}

func (p Property) IsRelation() bool {
	return len(p.Fields) > 0
}

func mergeProps(p1, p2 Property) Property {
	if p1.Index == 0 {
		p1.Index = p2.Index
	}
	if p1.Fields == nil {
		p1.Fields = p2.Fields
	} else {
		for k, v := range p2.Fields {
			p1.Fields[k] = mergeProps(p1.Fields[k], v)
		}
	}
	return p1
}

// ExtractProperties runs through the filter and returns a index -> Property map of the fields 
// being filtered on. 
func ExtractProperties(conditions map[connor.FilterKey]any) map[int]Property {
	properties := map[int]Property{}
	for k, v := range conditions {
		switch typedKey := k.(type) {
		case *mapper.PropertyIndex:
			prop := properties[typedKey.Index]
			prop.Index = typedKey.Index
			relatedProps := ExtractProperties(v.(map[connor.FilterKey]any))
			properties[typedKey.Index] = mergeProps(prop, Property{Fields: relatedProps})
		case *mapper.Operator:
			if typedKey.Operation == request.FilterOpAnd || typedKey.Operation == request.FilterOpOr {
				compoundContent := v.([]any)
				for _, compoundFilter := range compoundContent {
					props := ExtractProperties(compoundFilter.(map[connor.FilterKey]any))
					for _, prop := range props {
						existingProp := properties[prop.Index]
						properties[prop.Index] = mergeProps(existingProp, prop)
					}
				}
			} else if typedKey.Operation == request.FilterOpNot {
				props := ExtractProperties(v.(map[connor.FilterKey]any))
				for _, prop := range props {
					existingProp := properties[prop.Index]
					properties[prop.Index] = mergeProps(existingProp, prop)
				}
			}
		}
	}
	if len(properties) == 0 {
		return nil
	}
	return properties
}
