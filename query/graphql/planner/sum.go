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

func (p *Planner) Sum(
	sourceInfo *sourceInfo,
	field *parser.Field,
	parent *parser.Select,
) (*sumNode, error) {
	source, err := field.GetAggregateSource()
	if err != nil {
		return nil, err
	}

	sourceProperty := p.getSourceProperty(source, parent)
	isFloat, err := p.isValueFloat(sourceInfo, parent, source, sourceProperty)
	if err != nil {
		return nil, err
	}

	return &sumNode{
		p:                p,
		isFloat:          isFloat,
		sourceCollection: source.HostProperty,
		sourceProperty:   sourceProperty,
		virtualFieldId:   field.Name,
	}, nil
}

// Returns true if the value to be summed is a float, otherwise false.
func (p *Planner) isValueFloat(
	sourceInfo *sourceInfo,
	parent parser.Selection,
	source parser.AggregateTarget,
	sourceProperty string,
) (bool, error) {
	sourceFieldDescription, err := p.getSourceField(
		sourceInfo,
		parent,
		source,
		sourceProperty,
	)
	if err != nil {
		return false, err
	}

	return sourceFieldDescription.Kind == client.FieldKind_FLOAT_ARRAY ||
		sourceFieldDescription.Kind == client.FieldKind_FLOAT, nil
}

// Gets the root underlying field of the aggregate.
// This could be several layers deap if aggregating an aggregate.
func (p *Planner) getSourceField(
	sourceInfo *sourceInfo,
	parent parser.Selection,
	source parser.AggregateTarget,
	sourceProperty string,
) (client.FieldDescription, error) {
	if source.ChildProperty == "" {
		// If path length is one - we are summing an inline array
		fieldDescription, fieldDescriptionFound := sourceInfo.collectionDescription.GetField(source.HostProperty)
		if !fieldDescriptionFound {
			return client.FieldDescription{}, fmt.Errorf(
				"Unable to find field description for field: %s",
				source.HostProperty,
			)
		}
		return fieldDescription, nil
	}

	// If path length is two, we are summing a group or a child relationship
	if source.ChildProperty == parser.CountFieldName {
		// If we are summing a count, we know it is an int and can return early
		return client.FieldDescription{
			Kind: client.FieldKind_INT,
		}, nil
	}

	if _, isAggregate := parser.Aggregates[source.ChildProperty]; isAggregate {
		// If we are aggregating an aggregate, we need to traverse the aggregation chain down to
		// the root field in order to determine the value type.  This is recursive to allow handling
		// of N-depth aggregations (e.g. sum of sum of sum of...)
		var sourceField *parser.Field
		var sourceParent parser.Selection
		for _, field := range parent.GetSelections() {
			if field.GetName() == source.HostProperty {
				sourceParent = field
			}
		}

		for _, field := range sourceParent.GetSelections() {
			if field.GetAlias() == source.ChildProperty {
				sourceField = field.(*parser.Field)
				break
			}
		}
		sourceSource, err := sourceField.GetAggregateSource()
		if err != nil {
			return client.FieldDescription{}, err
		}

		sourceSourceProperty := p.getSourceProperty(sourceSource, sourceParent)
		return p.getSourceField(
			sourceInfo,
			sourceParent,
			sourceSource,
			sourceSourceProperty,
		)
	}

	if source.HostProperty == parser.GroupFieldName {
		// If the source collection is a group, then the description of the collection
		// to sum is this object.
		fieldDescription, fieldDescriptionFound := sourceInfo.collectionDescription.GetField(sourceProperty)
		if !fieldDescriptionFound {
			return client.FieldDescription{},
				fmt.Errorf("Unable to find field description for field: %s", sourceProperty)
		}
		return fieldDescription, nil
	}

	parentFieldDescription, parentFieldDescriptionFound := sourceInfo.collectionDescription.GetField(source.HostProperty)
	if !parentFieldDescriptionFound {
		return client.FieldDescription{}, fmt.Errorf(
			"Unable to find parent field description for field: %s",
			source.HostProperty,
		)
	}

	collectionDescription, err := p.getCollectionDesc(parentFieldDescription.Schema)
	if err != nil {
		return client.FieldDescription{}, err
	}

	fieldDescription, fieldDescriptionFound := collectionDescription.GetField(sourceProperty)
	if !fieldDescriptionFound {
		return client.FieldDescription{},
			fmt.Errorf("Unable to find child field description for field: %s", sourceProperty)
	}
	return fieldDescription, nil
}

// Gets the name of the immediate value-property to be aggregated.
func (p *Planner) getSourceProperty(source parser.AggregateTarget, parent parser.Selection) string {
	if source.ChildProperty == "" {
		return ""
	}

	if _, isAggregate := parser.Aggregates[source.ChildProperty]; isAggregate {
		for _, field := range parent.GetSelections() {
			if field.GetName() == source.HostProperty {
				for _, childField := range field.(*parser.Select).Fields {
					if childField.GetAlias() == source.ChildProperty {
						return childField.(*parser.Field).GetName()
					}
				}
			}
		}
	}

	return source.ChildProperty
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
					case int:
						sum += float64(v)
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
