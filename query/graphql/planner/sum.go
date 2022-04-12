// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

type sumNode struct {
	p    *Planner
	plan planNode

	isFloat          bool
	sourceCollection string
	sourceProperty   string
	virtualFieldId   string
}

func (p *Planner) Sum(sourceInfo *sourceInfo, field *parser.Field) (*sumNode, error) {
	var isFloat bool

	source, err := field.GetAggregateSource()
	if err != nil {
		return nil, err
	}

	sourceCollection := source[0]
	sourceProperty := getSourceProperty(source)

	if colDesc := sourceInfo.collectionDescription; len(source) == 1 {
		// If path length is one - we are summing an inline array
		fieldDescription, fieldDescriptionFound := colDesc.GetField(sourceCollection)
		if !fieldDescriptionFound {
			return nil, fmt.Errorf(
				"Unable to find field description for field: %s",
				sourceCollection,
			)
		}

		isFloat = fieldDescription.Kind == client.FieldKind_FLOAT_ARRAY
	} else if len(source) == 2 {
		// If path length is two, we are summing a group or a child relationship
		var childFieldDescription client.FieldDescription
		if sourceCollection == parser.GroupFieldName {
			// If the source collection is a group, then the description of the collection
			//  to sum is this object.
			fieldDescription, fieldDescriptionFound := colDesc.GetField(sourceProperty)
			if !fieldDescriptionFound {
				return nil, fmt.Errorf("Unable to find field description for field: %s", sourceProperty)
			}
			childFieldDescription = fieldDescription
		} else {
			parentFieldDescription, parentFieldDescriptionFound := colDesc.GetField(sourceCollection)
			if !parentFieldDescriptionFound {
				return nil, fmt.Errorf(
					"Unable to find parent field description for field: %s",
					sourceCollection,
				)
			}
			collectionDescription, err := p.getCollectionDesc(parentFieldDescription.Schema)
			if err != nil {
				return nil, err
			}
			fieldDescription, fieldDescriptionFound := collectionDescription.GetField(sourceProperty)
			if !fieldDescriptionFound {
				return nil, fmt.Errorf("Unable to find child field description for field: %s", sourceProperty)
			}
			childFieldDescription = fieldDescription
		}

		isFloat = childFieldDescription.Kind == client.FieldKind_FLOAT
	} else {
		return nil, fmt.Errorf("Sum must be provided with a property to sum.")
	}

	return &sumNode{
		p:                p,
		isFloat:          isFloat,
		sourceCollection: sourceCollection,
		sourceProperty:   sourceProperty,
		virtualFieldId:   field.Name,
	}, nil
}

func getSourceProperty(source []string) string {
	if len(source) == 1 {
		return ""
	}

	return source[1]
}

func (n *sumNode) Init() error {
	return n.plan.Init()
}

func (n *sumNode) Start() error           { return n.plan.Start() }
func (n *sumNode) Spans(spans core.Spans) { n.plan.Spans(spans) }
func (n *sumNode) Close() error           { return n.plan.Close() }
func (n *sumNode) Source() planNode       { return n.plan }

func (n *sumNode) Values() map[string]interface{} {
	value := n.plan.Values()

	sum := float64(0)

	if child, hasProperty := value[n.sourceCollection]; hasProperty {
		switch childCollection := child.(type) {
		case []map[string]interface{}:
			for _, childItem := range childCollection {
				if childProperty, hasChildProperty := childItem[n.sourceProperty]; hasChildProperty {
					switch v := childProperty.(type) {
					case int64:
						sum += float64(v)
					case uint64:
						sum += float64(v)
					case float64:
						sum += v
					default:
						// do nothing, cannot be summed
					}
				}
			}
		case []int64:
			for _, childItem := range childCollection {
				sum += float64(childItem)
			}
		case []float64:
			for _, childItem := range childCollection {
				sum += childItem
			}
		}
	}

	var typedSum interface{}
	if n.isFloat {
		typedSum = sum
	} else {
		typedSum = int64(sum)
	}
	value[n.virtualFieldId] = typedSum

	return value
}

func (n *sumNode) Next() (bool, error) {
	return n.plan.Next()
}

func (n *sumNode) SetPlan(p planNode) { n.plan = p }
